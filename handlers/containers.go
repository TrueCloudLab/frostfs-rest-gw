package handlers

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	containerv2 "github.com/TrueCloudLab/frostfs-api-go/v2/container"
	"github.com/TrueCloudLab/frostfs-api-go/v2/refs"
	sessionv2 "github.com/TrueCloudLab/frostfs-api-go/v2/session"
	"github.com/TrueCloudLab/frostfs-rest-gw/gen/models"
	"github.com/TrueCloudLab/frostfs-rest-gw/gen/restapi/operations"
	"github.com/TrueCloudLab/frostfs-rest-gw/internal/util"
	"github.com/TrueCloudLab/frostfs-sdk-go/container"
	"github.com/TrueCloudLab/frostfs-sdk-go/container/acl"
	cid "github.com/TrueCloudLab/frostfs-sdk-go/container/id"
	"github.com/TrueCloudLab/frostfs-sdk-go/netmap"
	"github.com/TrueCloudLab/frostfs-sdk-go/pool"
	"github.com/TrueCloudLab/frostfs-sdk-go/session"
	"github.com/TrueCloudLab/frostfs-sdk-go/user"
	"github.com/TrueCloudLab/frostfs-sdk-go/version"
	"github.com/go-openapi/runtime/middleware"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"go.uber.org/zap"
)

const (
	defaultPlacementPolicy = "REP 3"
	defaultBasicACL        = acl.NamePrivate
	attributeName          = "Name"
	attributeTimestamp     = "Timestamp"
)

// PutContainers handler that creates container in FrostFS.
func (a *API) PutContainers(params operations.PutContainerParams, principal *models.Principal) middleware.Responder {
	st, err := formSessionTokenFromHeaders(principal, params.XBearerSignature, params.XBearerSignatureKey, sessionv2.ContainerVerbPut)
	if err != nil {
		resp := a.logAndGetErrorResponse("invalid session token headers", err)
		return operations.NewPutContainerBadRequest().WithPayload(resp)
	}

	stoken, err := prepareSessionToken(st, *params.WalletConnect)
	if err != nil {
		resp := a.logAndGetErrorResponse("invalid session token", err)
		return operations.NewPutContainerBadRequest().WithPayload(resp)
	}

	cnrID, err := createContainer(params.HTTPRequest.Context(), a.pool, stoken, &params)
	if err != nil {
		resp := a.logAndGetErrorResponse("create container", err)
		return operations.NewPutContainerBadRequest().WithPayload(resp)
	}

	var resp operations.PutContainerOKBody
	resp.ContainerID = util.NewString(cnrID.EncodeToString())

	return operations.NewPutContainerOK().
		WithPayload(&resp).
		WithAccessControlAllowOrigin("*")
}

// GetContainer handler that returns container info.
func (a *API) GetContainer(params operations.GetContainerParams) middleware.Responder {
	cnrID, err := parseContainerID(params.ContainerID)
	if err != nil {
		resp := a.logAndGetErrorResponse("invalid container id", err)
		return operations.NewGetContainerBadRequest().WithPayload(resp)
	}

	cnrInfo, err := getContainerInfo(params.HTTPRequest.Context(), a.pool, cnrID)
	if err != nil {
		resp := a.logAndGetErrorResponse("get container", err)
		return operations.NewGetContainerBadRequest().WithPayload(resp)
	}

	return operations.NewGetContainerOK().
		WithPayload(cnrInfo).
		WithAccessControlAllowOrigin("*")
}

// PutContainerEACL handler that update container eacl.
func (a *API) PutContainerEACL(params operations.PutContainerEACLParams, principal *models.Principal) middleware.Responder {
	cnrID, err := parseContainerID(params.ContainerID)
	if err != nil {
		resp := a.logAndGetErrorResponse("invalid container id", err)
		return operations.NewPutContainerEACLBadRequest().WithPayload(resp)
	}

	if err = checkContainerExtendable(params.HTTPRequest.Context(), a.pool, cnrID); err != nil {
		resp := a.logAndGetErrorResponse("check acl allowance", err)
		return operations.NewPutContainerEACLBadRequest().WithPayload(resp)
	}

	st, err := formSessionTokenFromHeaders(principal, params.XBearerSignature, params.XBearerSignatureKey, sessionv2.ContainerVerbSetEACL)
	if err != nil {
		resp := a.logAndGetErrorResponse("invalid session token headers", err)
		return operations.NewPutContainerEACLBadRequest().WithPayload(resp)
	}

	stoken, err := prepareSessionToken(st, *params.WalletConnect)
	if err != nil {
		resp := a.logAndGetErrorResponse("invalid session token", err)
		return operations.NewPutContainerEACLBadRequest().WithPayload(resp)
	}

	if err = setContainerEACL(params.HTTPRequest.Context(), a.pool, cnrID, stoken, params.Eacl); err != nil {
		resp := a.logAndGetErrorResponse("failed set container eacl", err)
		return operations.NewPutContainerEACLBadRequest().WithPayload(resp)
	}

	return operations.NewPutContainerEACLOK().
		WithPayload(util.NewSuccessResponse()).
		WithAccessControlAllowOrigin("*")
}

// GetContainerEACL handler that returns container eacl.
func (a *API) GetContainerEACL(params operations.GetContainerEACLParams) middleware.Responder {
	cnrID, err := parseContainerID(params.ContainerID)
	if err != nil {
		resp := a.logAndGetErrorResponse("invalid container id", err)
		return operations.NewGetContainerEACLBadRequest().WithPayload(resp)
	}

	resp, err := getContainerEACL(params.HTTPRequest.Context(), a.pool, cnrID)
	if err != nil {
		errResp := a.logAndGetErrorResponse("failed to get container eacl", err)
		return operations.NewGetContainerEACLBadRequest().WithPayload(errResp)
	}

	return operations.NewGetContainerEACLOK().
		WithPayload(resp).
		WithAccessControlAllowOrigin("*")
}

// ListContainer handler that returns containers.
func (a *API) ListContainer(params operations.ListContainersParams) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	var ownerID user.ID
	if err := ownerID.DecodeString(params.OwnerID); err != nil {
		resp := a.logAndGetErrorResponse("invalid owner id", err)
		return operations.NewListContainersBadRequest().WithPayload(resp)
	}

	var prm pool.PrmContainerList
	prm.SetOwnerID(ownerID)

	ids, err := a.pool.ListContainers(ctx, prm)
	if err != nil {
		resp := a.logAndGetErrorResponse("list containers", err)
		return operations.NewListContainersBadRequest().WithPayload(resp)
	}

	offset := int(*params.Offset)
	size := int(*params.Limit)

	if offset > len(ids)-1 {
		res := &models.ContainerList{
			Size:       util.NewInteger(0),
			Containers: []*models.ContainerInfo{},
		}
		return operations.NewListContainersOK().
			WithPayload(res).
			WithAccessControlAllowOrigin("*")
	}

	if offset+size > len(ids) {
		size = len(ids) - offset
	}

	res := &models.ContainerList{
		Size:       util.NewInteger(int64(size)),
		Containers: make([]*models.ContainerInfo, 0, size),
	}

	for _, id := range ids[offset : offset+size] {
		cnrInfo, err := getContainerInfo(ctx, a.pool, id)
		if err != nil {
			resp := a.logAndGetErrorResponse("get container", err, zap.String("cid", id.String()))
			return operations.NewListContainersBadRequest().WithPayload(resp)
		}
		res.Containers = append(res.Containers, cnrInfo)
	}

	return operations.NewListContainersOK().
		WithPayload(res).
		WithAccessControlAllowOrigin("*")
}

// DeleteContainer handler that returns container info.
func (a *API) DeleteContainer(params operations.DeleteContainerParams, principal *models.Principal) middleware.Responder {
	st, err := formSessionTokenFromHeaders(principal, params.XBearerSignature, params.XBearerSignatureKey, sessionv2.ContainerVerbDelete)
	if err != nil {
		resp := a.logAndGetErrorResponse("invalid session token headers", err)
		return operations.NewDeleteContainerBadRequest().WithPayload(resp)
	}

	stoken, err := prepareSessionToken(st, *params.WalletConnect)
	if err != nil {
		resp := a.logAndGetErrorResponse("invalid session token", err)
		return operations.NewDeleteContainerBadRequest().WithPayload(resp)
	}

	cnrID, err := parseContainerID(params.ContainerID)
	if err != nil {
		resp := a.logAndGetErrorResponse("invalid container id", err)
		return operations.NewDeleteContainerBadRequest().WithPayload(resp)
	}

	var prm pool.PrmContainerDelete
	prm.SetContainerID(cnrID)
	prm.SetSessionToken(stoken)

	if err = a.pool.DeleteContainer(params.HTTPRequest.Context(), prm); err != nil {
		resp := a.logAndGetErrorResponse("delete container", err, zap.String("container", params.ContainerID))
		return operations.NewDeleteContainerBadRequest().WithPayload(resp)
	}

	return operations.NewDeleteContainerOK().
		WithPayload(util.NewSuccessResponse()).
		WithAccessControlAllowOrigin("*")
}

func checkContainerExtendable(ctx context.Context, p *pool.Pool, cnrID cid.ID) error {
	cnr, err := getContainer(ctx, p, cnrID)
	if err != nil {
		return fmt.Errorf("get container: %w", err)
	}

	if !cnr.BasicACL().Extendable() {
		return fmt.Errorf("container acl isn't extendable")
	}

	return nil
}

func getContainer(ctx context.Context, p *pool.Pool, cnrID cid.ID) (container.Container, error) {
	var prm pool.PrmContainerGet
	prm.SetContainerID(cnrID)

	return p.GetContainer(ctx, prm)
}

func getContainerInfo(ctx context.Context, p *pool.Pool, cnrID cid.ID) (*models.ContainerInfo, error) {
	cnr, err := getContainer(ctx, p, cnrID)
	if err != nil {
		return nil, err
	}

	var attrs []*models.Attribute
	cnr.IterateAttributes(func(key, val string) {
		attrs = append(attrs, &models.Attribute{
			Key:   util.NewString(key),
			Value: util.NewString(val),
		})
	})

	var sb strings.Builder
	if err = cnr.PlacementPolicy().WriteStringTo(&sb); err != nil {
		return nil, fmt.Errorf("writer policy to string: %w", err)
	}

	return &models.ContainerInfo{
		ContainerID:     util.NewString(cnrID.String()),
		ContainerName:   util.NewString(container.Name(cnr)),
		OwnerID:         util.NewString(cnr.Owner().String()),
		BasicACL:        util.NewString(cnr.BasicACL().EncodeToString()),
		CannedACL:       friendlyBasicACL(cnr.BasicACL()),
		PlacementPolicy: util.NewString(sb.String()),
		Attributes:      attrs,
		Version:         util.NewString(getContainerVersion(cnr).String()),
	}, nil
}

func friendlyBasicACL(basicACL acl.Basic) string {
	switch basicACL {
	case acl.Private:
		return acl.NamePrivate
	case acl.PrivateExtended:
		return acl.NamePrivateExtended
	case acl.PublicRO:
		return acl.NamePublicRO
	case acl.PublicROExtended:
		return acl.NamePublicROExtended
	case acl.PublicRW:
		return acl.NamePublicRW
	case acl.PublicRWExtended:
		return acl.NamePublicRWExtended
	case acl.PublicAppend:
		return acl.NamePublicAppend
	case acl.PublicAppendExtended:
		return acl.NamePublicAppendExtended
	default:
		return ""
	}
}

func getContainerVersion(cnr container.Container) version.Version {
	var v2cnr containerv2.Container
	cnr.WriteToV2(&v2cnr)

	var cnrVersion version.Version
	v2version := v2cnr.GetVersion()
	if v2version != nil {
		cnrVersion = version.Version(*v2version)
	}

	return cnrVersion
}

func parseContainerID(containerID string) (cid.ID, error) {
	var cnrID cid.ID
	if err := cnrID.DecodeString(containerID); err != nil {
		return cid.ID{}, fmt.Errorf("parse container id '%s': %w", containerID, err)
	}

	return cnrID, nil
}

func setContainerEACL(ctx context.Context, p *pool.Pool, cnrID cid.ID, stoken session.Container, eaclPrm *models.Eacl) error {
	table, err := util.ToNativeTable(eaclPrm.Records)
	if err != nil {
		return err
	}

	table.SetCID(cnrID)

	var prm pool.PrmContainerSetEACL
	prm.SetTable(*table)
	prm.WithinSession(stoken)

	return p.SetEACL(ctx, prm)
}

func getContainerEACL(ctx context.Context, p *pool.Pool, cnrID cid.ID) (*models.Eacl, error) {
	var prm pool.PrmContainerEACL
	prm.SetContainerID(cnrID)

	table, err := p.GetEACL(ctx, prm)
	if err != nil {
		return nil, fmt.Errorf("get eacl: %w", err)
	}

	tableResp := &models.Eacl{
		ContainerID: cnrID.EncodeToString(),
		Records:     make([]*models.Record, len(table.Records())),
	}

	for i, rec := range table.Records() {
		record, err := util.FromNativeRecord(rec)
		if err != nil {
			return nil, fmt.Errorf("couldn't transform record from native: %w", err)
		}
		tableResp.Records[i] = record
	}

	return tableResp, nil
}

func createContainer(ctx context.Context, p *pool.Pool, stoken session.Container, params *operations.PutContainerParams) (cid.ID, error) {
	request := params.Container

	if request.PlacementPolicy == "" {
		request.PlacementPolicy = defaultPlacementPolicy
	}
	var policy netmap.PlacementPolicy
	err := policy.DecodeString(request.PlacementPolicy)
	if err != nil {
		return cid.ID{}, fmt.Errorf("couldn't parse placement policy: %w", err)
	}

	if request.BasicACL == "" {
		request.BasicACL = defaultBasicACL
	}

	basicACL, err := decodeBasicACL(request.BasicACL)
	if err != nil {
		return cid.ID{}, fmt.Errorf("couldn't parse basic acl: %w", err)
	}

	var cnr container.Container
	cnr.Init()
	cnr.SetPlacementPolicy(policy)
	cnr.SetBasicACL(basicACL)
	cnr.SetOwner(stoken.Issuer())

	container.SetCreationTime(&cnr, time.Now())

	if request.ContainerName != "" {
		container.SetName(&cnr, request.ContainerName)
	}

	for _, attr := range request.Attributes {
		switch *attr.Key {
		case attributeName, attributeTimestamp,
			containerv2.SysAttributeName, containerv2.SysAttributeZone:
		default:
			cnr.SetAttribute(*attr.Key, *attr.Value)
		}
	}

	if *params.NameScopeGlobal { // we don't check for nil because there is default false value
		if err = checkNNSContainerName(request.ContainerName); err != nil {
			return cid.ID{}, fmt.Errorf("invalid container name: %w", err)
		}

		var domain container.Domain
		domain.SetName(request.ContainerName)
		container.WriteDomain(&cnr, domain)
	}

	var prm pool.PrmContainerPut
	prm.SetContainer(cnr)
	prm.WithinSession(stoken)

	cnrID, err := p.PutContainer(ctx, prm)
	if err != nil {
		return cid.ID{}, fmt.Errorf("put container: %w", err)
	}

	return cnrID, nil
}

func checkNNSContainerName(name string) error {
	length := len(name)
	if length < 3 || 255 < length {
		return fmt.Errorf("invalid length: %d", length)
	}
	fragments := strings.Split(name, ".")

	for _, fragment := range fragments {
		fLength := len(fragment)
		if fLength < 1 || 63 < fLength {
			return fmt.Errorf("invalid fragment length: %d", fLength)
		}

		if !isAlNum(fragment[0]) || !isAlNum(fragment[fLength-1]) {
			return fmt.Errorf("invalid fragment: '%s'", fragment)
		}

		for i := 1; i < fLength-1; i++ {
			if fragment[i] != '-' && !isAlNum(fragment[i]) {
				return fmt.Errorf("invalid fragment: '%s'", fragment)
			}
		}
	}

	return nil
}

func isAlNum(c uint8) bool {
	return c >= 'a' && c <= 'z' || c >= '0' && c <= '9'
}

func prepareSessionToken(st *SessionToken, isWalletConnect bool) (session.Container, error) {
	data, err := base64.StdEncoding.DecodeString(st.Token)
	if err != nil {
		return session.Container{}, fmt.Errorf("can't base64-decode session token: %w", err)
	}

	signature, err := hex.DecodeString(st.Signature)
	if err != nil {
		return session.Container{}, fmt.Errorf("couldn't decode signature: %w", err)
	}

	ownerKey, err := keys.NewPublicKeyFromString(st.Key)
	if err != nil {
		return session.Container{}, fmt.Errorf("couldn't fetch session token owner key: %w", err)
	}

	body := new(sessionv2.TokenBody)
	if err = body.Unmarshal(data); err != nil {
		return session.Container{}, fmt.Errorf("can't unmarshal session token: %w", err)
	}

	if sessionContext, ok := body.GetContext().(*sessionv2.ContainerSessionContext); !ok {
		return session.Container{}, errors.New("expected container session context but got something different")
	} else if sessionContext.Verb() != st.Verb {
		return session.Container{}, fmt.Errorf("invalid container session verb '%s', expected: '%s'", sessionContext.Verb().String(), st.Verb.String())
	}

	v2signature := new(refs.Signature)
	v2signature.SetScheme(refs.ECDSA_SHA512)
	if isWalletConnect {
		v2signature.SetScheme(refs.ECDSA_RFC6979_SHA256_WALLET_CONNECT)
	}
	v2signature.SetSign(signature)
	v2signature.SetKey(ownerKey.Bytes())

	var v2token sessionv2.Token
	v2token.SetBody(body)
	v2token.SetSignature(v2signature)

	var stoken session.Container
	if err = stoken.ReadFromV2(v2token); err != nil {
		return session.Container{}, fmt.Errorf("read from v2 token: %w", err)
	}

	if !stoken.VerifySignature() {
		return session.Container{}, fmt.Errorf("invalid signature")
	}

	return stoken, err
}
