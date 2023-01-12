package handlers

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"

	objectv2 "github.com/TrueCloudLab/frostfs-api-go/v2/object"
	"github.com/TrueCloudLab/frostfs-rest-gw/gen/models"
	"github.com/TrueCloudLab/frostfs-rest-gw/gen/restapi/operations"
	"github.com/TrueCloudLab/frostfs-rest-gw/internal/util"
	"github.com/TrueCloudLab/frostfs-sdk-go/bearer"
	"github.com/TrueCloudLab/frostfs-sdk-go/checksum"
	"github.com/TrueCloudLab/frostfs-sdk-go/container"
	cid "github.com/TrueCloudLab/frostfs-sdk-go/container/id"
	"github.com/TrueCloudLab/frostfs-sdk-go/object"
	oid "github.com/TrueCloudLab/frostfs-sdk-go/object/id"
	"github.com/TrueCloudLab/frostfs-sdk-go/pool"
	"github.com/TrueCloudLab/frostfs-sdk-go/storagegroup"
	"github.com/TrueCloudLab/tzhash/tz"
	"github.com/go-openapi/runtime/middleware"
)

// PutStorageGroup handler that create a new storage group.
func (a *API) PutStorageGroup(params operations.PutStorageGroupParams, principal *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	cnrID, err := parseContainerID(params.ContainerID)
	if err != nil {
		resp := a.logAndGetErrorResponse("invalid container id", err)
		return operations.NewPutStorageGroupBadRequest().WithPayload(resp)
	}

	btoken, err := getBearerToken(principal, params.XBearerSignature, params.XBearerSignatureKey, *params.WalletConnect, *params.FullBearer)
	if err != nil {
		resp := a.logAndGetErrorResponse("invalid bearer token", err)
		return operations.NewPutStorageGroupBadRequest().WithPayload(resp)
	}

	sg, err := a.formStorageGroup(ctx, cnrID, btoken, params.StorageGroup)
	if err != nil {
		resp := a.logAndGetErrorResponse("form storage group", err)
		return operations.NewPutStorageGroupBadRequest().WithPayload(resp)
	}

	objID, err := a.putStorageGroupObject(ctx, cnrID, btoken, params.StorageGroup.Name, *sg)
	if err != nil {
		resp := a.logAndGetErrorResponse("put storage group", err)
		return operations.NewPutStorageGroupBadRequest().WithPayload(resp)
	}

	var resp models.Address
	resp.ContainerID = util.NewString(params.ContainerID)
	resp.ObjectID = util.NewString(objID.String())

	return operations.NewPutStorageGroupOK().WithPayload(&resp)
}

// ListStorageGroups handler that create a new storage group.
func (a *API) ListStorageGroups(params operations.ListStorageGroupsParams, principal *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	cnrID, err := parseContainerID(params.ContainerID)
	if err != nil {
		resp := a.logAndGetErrorResponse("invalid container id", err)
		return operations.NewListStorageGroupsBadRequest().WithPayload(resp)
	}

	btoken, err := getBearerToken(principal, params.XBearerSignature, params.XBearerSignatureKey, *params.WalletConnect, *params.FullBearer)
	if err != nil {
		resp := a.logAndGetErrorResponse("invalid bearer token", err)
		return operations.NewListStorageGroupsBadRequest().WithPayload(resp)
	}

	var filters object.SearchFilters
	filters.AddTypeFilter(object.MatchStringEqual, object.TypeStorageGroup)

	var prm pool.PrmObjectSearch
	prm.SetContainerID(cnrID)
	attachBearer(&prm, btoken)
	prm.SetFilters(filters)

	resSearch, err := a.pool.SearchObjects(ctx, prm)
	if err != nil {
		resp := a.logAndGetErrorResponse("failed to search objects", err)
		return operations.NewListStorageGroupsBadRequest().WithPayload(resp)
	}

	var iterateErr error
	var sgInfo *models.StorageGroupBaseInfo
	var storageGroups []*models.StorageGroupBaseInfo

	err = resSearch.Iterate(func(id oid.ID) bool {
		if sgInfo, iterateErr = headObjectStorageGroupBaseInfo(ctx, a.pool, cnrID, id, btoken); iterateErr != nil {
			return true
		}

		storageGroups = append(storageGroups, sgInfo)
		return false
	})
	if err == nil {
		err = iterateErr
	}
	if err != nil {
		resp := a.logAndGetErrorResponse("failed to search storage groups", err)
		return operations.NewListStorageGroupsBadRequest().WithPayload(resp)
	}

	resp := &models.StorageGroupList{
		Size:          util.NewInteger(int64(len(storageGroups))),
		StorageGroups: storageGroups,
	}

	return operations.NewListStorageGroupsOK().WithPayload(resp)
}

func headObjectStorageGroupBaseInfo(ctx context.Context, p *pool.Pool, cnrID cid.ID, objID oid.ID, btoken *bearer.Token) (*models.StorageGroupBaseInfo, error) {
	var addr oid.Address
	addr.SetContainer(cnrID)
	addr.SetObject(objID)

	var prm pool.PrmObjectHead
	prm.SetAddress(addr)
	attachBearer(&prm, btoken)

	objInfo, err := p.HeadObject(ctx, prm)
	if err != nil {
		return nil, fmt.Errorf("head object '%s': %w", objID.EncodeToString(), err)
	}

	resp := &models.StorageGroupBaseInfo{
		Address: &models.Address{
			ContainerID: util.NewString(cnrID.String()),
			ObjectID:    util.NewString(objID.String()),
		},
	}

	expEpoch := "0"
	for _, attr := range objInfo.Attributes() {
		switch attr.Key() {
		case object.AttributeFileName:
			resp.Name = attr.Value()
		case objectv2.SysAttributeExpEpoch:
			if _, err = strconv.ParseUint(attr.Value(), 10, 64); err != nil {
				return nil, fmt.Errorf("invalid expiration epoch '%s': %w", attr.Value(), err)
			}
			expEpoch = attr.Value()
		}
	}

	resp.ExpirationEpoch = &expEpoch

	return resp, nil
}

// DeleteStorageGroup handler that removes storage group from NeoFS.
func (a *API) DeleteStorageGroup(params operations.DeleteStorageGroupParams, principal *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	addr, err := parseAddress(params.ContainerID, params.StorageGroupID)
	if err != nil {
		resp := a.logAndGetErrorResponse("invalid address", err)
		return operations.NewDeleteStorageGroupBadRequest().WithPayload(resp)
	}

	btoken, err := getBearerToken(principal, params.XBearerSignature, params.XBearerSignatureKey, *params.WalletConnect, *params.FullBearer)
	if err != nil {
		resp := a.logAndGetErrorResponse("failed to get bearer token", err)
		return operations.NewDeleteStorageGroupBadRequest().WithPayload(resp)
	}

	var prm pool.PrmObjectDelete
	prm.SetAddress(addr)
	attachBearer(&prm, btoken)

	if err = a.pool.DeleteObject(ctx, prm); err != nil {
		resp := a.logAndGetErrorResponse("failed to delete storage group", err)
		return operations.NewDeleteStorageGroupBadRequest().WithPayload(resp)
	}

	return operations.NewDeleteStorageGroupOK().WithPayload(util.NewSuccessResponse())
}

// GetStorageGroup handler that get storage group info.
func (a *API) GetStorageGroup(params operations.GetStorageGroupParams, principal *models.Principal) middleware.Responder {
	errorResponse := operations.NewGetObjectInfoBadRequest()
	ctx := params.HTTPRequest.Context()

	addr, err := parseAddress(params.ContainerID, params.StorageGroupID)
	if err != nil {
		resp := a.logAndGetErrorResponse("invalid address", err)
		return errorResponse.WithPayload(resp)
	}

	btoken, err := getBearerToken(principal, params.XBearerSignature, params.XBearerSignatureKey, *params.WalletConnect, *params.FullBearer)
	if err != nil {
		resp := a.logAndGetErrorResponse("get bearer token", err)
		return errorResponse.WithPayload(resp)
	}

	var prm pool.PrmObjectGet
	prm.SetAddress(addr)
	attachBearer(&prm, btoken)

	objRes, err := a.pool.GetObject(ctx, prm)
	if err != nil {
		resp := a.logAndGetErrorResponse("get storage group object", err)
		return errorResponse.WithPayload(resp)
	}

	sb, err := a.readStorageGroup(objRes)
	if err != nil {
		resp := a.logAndGetErrorResponse("read storage group", err)
		return errorResponse.WithPayload(resp)
	}

	var sbHash string
	cs, ok := sb.ValidationDataHash()
	if ok {
		sbHash = hex.EncodeToString(cs.Value())
	}

	members := make([]string, len(sb.Members()))
	for i, objID := range sb.Members() {
		members[i] = objID.EncodeToString()
	}

	resp := &models.StorageGroup{
		Address: &models.Address{
			ContainerID: util.NewString(addr.Container().String()),
			ObjectID:    util.NewString(addr.Object().String()),
		},
		ExpirationEpoch: util.NewString(strconv.FormatUint(sb.ExpirationEpoch(), 10)),
		Size:            util.NewString(strconv.FormatUint(sb.ValidationDataSize(), 10)),
		Hash:            sbHash,
		Members:         members,
		Name:            getStorageGroupName(objRes.Header),
	}

	return operations.NewGetStorageGroupOK().WithPayload(resp)
}

func getStorageGroupName(obj object.Object) string {
	for _, attribute := range obj.Attributes() {
		if attribute.Key() == object.AttributeFileName {
			return attribute.Value()
		}
	}
	return ""
}

func (a *API) readStorageGroup(objRes pool.ResGetObject) (*storagegroup.StorageGroup, error) {
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, objRes.Payload); err != nil {
		return nil, fmt.Errorf("failed to copy storage group payload: %w", err)
	}

	obj := objRes.Header
	obj.SetPayload(buf.Bytes())

	var sb storagegroup.StorageGroup
	if err := storagegroup.ReadFromObject(&sb, obj); err != nil {
		return nil, fmt.Errorf("read storage group from object: %w", err)
	}

	return &sb, nil
}

func (a *API) formStorageGroup(ctx context.Context, cnrID cid.ID, btoken *bearer.Token, storageGroup *models.StorageGroupPutBody) (*storagegroup.StorageGroup, error) {
	members, err := a.parseStorageGroupMembers(storageGroup)
	if err != nil {
		return nil, fmt.Errorf("parse storage group members: %w", err)
	}

	hashDisabled, err := isHomomorphicHashingDisabled(ctx, a.pool, cnrID)
	if err != nil {
		return nil, fmt.Errorf("check if homomorphic hash disabled: %w", err)
	}

	sgSize, cs, err := a.getStorageGroupSizeAndHash(ctx, cnrID, btoken, members, !hashDisabled)
	if err != nil {
		return nil, fmt.Errorf("get storage group size: %w", err)
	}

	networkInfo, err := a.pool.NetworkInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("get network info: %w", err)
	}

	var sg storagegroup.StorageGroup
	sg.SetMembers(members)
	sg.SetValidationDataSize(sgSize)
	sg.SetExpirationEpoch(networkInfo.CurrentEpoch() + uint64(*storageGroup.Lifetime))

	if !hashDisabled {
		sg.SetValidationDataHash(*cs)
	}

	return &sg, nil
}

func (a *API) putStorageGroupObject(ctx context.Context, cnrID cid.ID, btoken *bearer.Token, fileName string, sg storagegroup.StorageGroup) (*oid.ID, error) {
	var attrFileName object.Attribute
	attrFileName.SetKey(object.AttributeFileName)
	attrFileName.SetValue(fileName)

	obj := object.New()
	obj.SetContainerID(cnrID)
	attachOwner(obj, btoken)
	obj.SetAttributes(attrFileName)

	storagegroup.WriteToObject(sg, obj)

	var prmPut pool.PrmObjectPut
	prmPut.SetHeader(*obj)
	attachBearer(&prmPut, btoken)

	objID, err := a.pool.PutObject(ctx, prmPut)
	if err != nil {
		return nil, fmt.Errorf("put object: %w", err)
	}

	return &objID, nil
}

func (a *API) getStorageGroupSizeAndHash(ctx context.Context, cnrID cid.ID, btoken *bearer.Token, members []oid.ID, needCalcHash bool) (uint64, *checksum.Checksum, error) {
	var (
		sgSize    uint64
		objHashes [][]byte
		addr      oid.Address
		prm       pool.PrmObjectHead
	)

	addr.SetContainer(cnrID)
	attachBearer(&prm, btoken)

	for _, objID := range members {
		addr.SetObject(objID)
		prm.SetAddress(addr)

		objInfo, err := a.pool.HeadObject(ctx, prm)
		if err != nil {
			return 0, nil, fmt.Errorf("chead object from storage group members, id '%s': %w", objID.EncodeToString(), err)
		}

		sgSize += objInfo.PayloadSize()

		if needCalcHash {
			cs, _ := objInfo.PayloadHomomorphicHash()
			objHashes = append(objHashes, cs.Value())
		}
	}

	if needCalcHash {
		sumHash, err := tz.Concat(objHashes)
		if err != nil {
			return 0, nil, fmt.Errorf("concat tz hashes: %w", err)
		}

		var cs checksum.Checksum
		tzHash := [64]byte{}
		copy(tzHash[:], sumHash)
		cs.SetTillichZemor(tzHash)

		return sgSize, &cs, nil
	}

	return sgSize, nil, nil
}

func (a *API) parseStorageGroupMembers(storageGroup *models.StorageGroupPutBody) ([]oid.ID, error) {
	var err error

	members := make([]oid.ID, len(storageGroup.Members))
	uniqueFilter := make(map[oid.ID]struct{}, len(members))

	for i, objIDStr := range storageGroup.Members {
		if err = members[i].DecodeString(objIDStr); err != nil {
			return nil, fmt.Errorf("invalid object id '%s': %w", objIDStr, err)
		}
		if _, ok := uniqueFilter[members[i]]; ok {
			return nil, fmt.Errorf("invalid storage group members: duplicate id '%s': %w", objIDStr, err)
		}
		uniqueFilter[members[i]] = struct{}{}
	}

	return members, nil
}

func isHomomorphicHashingDisabled(ctx context.Context, p *pool.Pool, cnrID cid.ID) (bool, error) {
	var prm pool.PrmContainerGet
	prm.SetContainerID(cnrID)

	cnr, err := p.GetContainer(ctx, prm)
	if err != nil {
		return false, fmt.Errorf("get container: %w", err)
	}

	return container.IsHomomorphicHashingDisabled(cnr), nil
}
