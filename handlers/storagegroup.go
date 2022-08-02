package handlers

import (
	"context"
	"fmt"

	"github.com/go-openapi/runtime/middleware"
	"github.com/nspcc-dev/neofs-rest-gw/gen/models"
	"github.com/nspcc-dev/neofs-rest-gw/gen/restapi/operations"
	"github.com/nspcc-dev/neofs-rest-gw/internal/util"
	"github.com/nspcc-dev/neofs-sdk-go/bearer"
	"github.com/nspcc-dev/neofs-sdk-go/checksum"
	"github.com/nspcc-dev/neofs-sdk-go/container"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/object"
	oid "github.com/nspcc-dev/neofs-sdk-go/object/id"
	"github.com/nspcc-dev/neofs-sdk-go/pool"
	"github.com/nspcc-dev/neofs-sdk-go/storagegroup"
	"github.com/nspcc-dev/tzhash/tz"
)

// PutStorageGroup handler that create a new storage group.
func (a *API) PutStorageGroup(params operations.PutStorageGroupParams, principal *models.Principal) middleware.Responder {
	ctx := params.HTTPRequest.Context()

	cnrID, err := parseContainerID(params.ContainerID)
	if err != nil {
		resp := a.logAndGetErrorResponse("invalid container id", err)
		return operations.NewPutStorageGroupBadRequest().WithPayload(resp)
	}

	btoken, err := getBearerToken(principal, params.XBearerSignature, params.XBearerSignatureKey, *params.WalletConnect)
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

func (a *API) formStorageGroup(ctx context.Context, cnrID cid.ID, btoken bearer.Token, storageGroup *models.StorageGroup) (*storagegroup.StorageGroup, error) {
	members, err := a.parseStorageGroupMembers(storageGroup)
	if err != nil {
		return nil, fmt.Errorf("parse storage group members: %w", err)
	}

	needCalcHash, err := isHomomorphicHashingDisabled(ctx, a.pool, cnrID)
	if err != nil {
		return nil, fmt.Errorf("check if homomorphic hash disabled: %w", err)
	}

	sgSize, cs, err := a.getStorageGroupSizeAndHash(ctx, cnrID, btoken, members, needCalcHash)
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

	if needCalcHash {
		sg.SetValidationDataHash(*cs)
	}

	return &sg, nil
}

func (a *API) putStorageGroupObject(ctx context.Context, cnrID cid.ID, btoken bearer.Token, fileName string, sg storagegroup.StorageGroup) (*oid.ID, error) {
	owner := bearer.ResolveIssuer(btoken)

	var attrFileName object.Attribute
	attrFileName.SetKey(object.AttributeFileName)
	attrFileName.SetValue(fileName)

	obj := object.New()
	obj.SetContainerID(cnrID)
	obj.SetOwnerID(&owner)
	obj.SetAttributes(attrFileName)

	storagegroup.WriteToObject(sg, obj)

	var prmPut pool.PrmObjectPut
	prmPut.SetHeader(*obj)
	prmPut.UseBearer(btoken)

	objID, err := a.pool.PutObject(ctx, prmPut)
	if err != nil {
		return nil, fmt.Errorf("put object: %w", err)
	}

	return objID, nil
}

func (a *API) getStorageGroupSizeAndHash(ctx context.Context, cnrID cid.ID, btoken bearer.Token, members []oid.ID, needCalcHash bool) (uint64, *checksum.Checksum, error) {
	var (
		sgSize    uint64
		objHashes [][]byte
		addr      oid.Address
		prm       pool.PrmObjectHead
	)

	addr.SetContainer(cnrID)
	prm.UseBearer(btoken)

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

func (a *API) parseStorageGroupMembers(storageGroup *models.StorageGroup) ([]oid.ID, error) {
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

	return container.IsHomomorphicHashingDisabled(*cnr), nil
}
