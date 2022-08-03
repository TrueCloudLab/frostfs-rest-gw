package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/TrueCloudLab/frostfs-rest-gw/gen/models"
	"github.com/TrueCloudLab/frostfs-rest-gw/gen/restapi"
	"github.com/TrueCloudLab/frostfs-rest-gw/gen/restapi/operations"
	"github.com/TrueCloudLab/frostfs-rest-gw/handlers"
	"github.com/TrueCloudLab/frostfs-rest-gw/internal/util"
	"github.com/TrueCloudLab/frostfs-sdk-go/bearer"
	"github.com/TrueCloudLab/frostfs-sdk-go/container"
	"github.com/TrueCloudLab/frostfs-sdk-go/container/acl"
	cid "github.com/TrueCloudLab/frostfs-sdk-go/container/id"
	frostfsecdsa "github.com/TrueCloudLab/frostfs-sdk-go/crypto/ecdsa"
	"github.com/TrueCloudLab/frostfs-sdk-go/eacl"
	"github.com/TrueCloudLab/frostfs-sdk-go/netmap"
	"github.com/TrueCloudLab/frostfs-sdk-go/object"
	oid "github.com/TrueCloudLab/frostfs-sdk-go/object/id"
	"github.com/TrueCloudLab/frostfs-sdk-go/pool"
	"github.com/TrueCloudLab/frostfs-sdk-go/user"
	"github.com/go-openapi/loads"
	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	devenvPrivateKey  = "1dd37fba80fec4e6a6f13fd708d8dcb3b29def768017052f6c930fa1c5d90bbb"
	testListenAddress = "localhost:8082"
	testHost          = "http://" + testListenAddress
	testContainerNode = "localhost:8080"
	testLocalNode     = "s01.frostfs.devenv:8080"
	containerName     = "test-container"
	localVersion      = "local"

	walletConnectQuery = "walletConnect"
	fullBearerQuery    = "fullBearer"
	// XBearerSignature header contains base64 encoded signature of the token body.
	XBearerSignature = "X-Bearer-Signature"
	// XBearerSignatureKey header contains hex encoded public key that corresponds the signature of the token body.
	XBearerSignatureKey = "X-Bearer-Signature-Key"
	// XBearerOwnerID header contains owner id (wallet address) that corresponds the signature of the token body.
	XBearerOwnerID = "X-Bearer-Owner-Id"
	// XBearerForAllUsers header specifies if we want all users can use token or only specific gate.
	XBearerForAllUsers = "X-Bearer-For-All-Users"

	// tests configuration.
	useWalletConnect    = true
	useLocalEnvironment = false
)

func TestIntegration(t *testing.T) {
	ctx := context.Background()
	key, err := keys.NewPrivateKeyFromHex(devenvPrivateKey)
	require.NoError(t, err)

	if useLocalEnvironment {
		runLocalTests(ctx, t, key)
	} else {
		runTestInContainer(ctx, t, key)
	}
}

func runLocalTests(ctx context.Context, t *testing.T, key *keys.PrivateKey) {
	runTests(ctx, t, key, localVersion)
}

func runTestInContainer(rootCtx context.Context, t *testing.T, key *keys.PrivateKey) {
	aioImage := "nspccdev/neofs-aio-testcontainer:"
	versions := []string{
		"0.29.0",
		"0.30.0",
		"0.32.0",
		"latest",
	}

	for _, version := range versions {
		ctx, cancel := context.WithCancel(rootCtx)
		aioContainer := createDockerContainer(ctx, t, aioImage+version)

		runTests(ctx, t, key, version)

		err := aioContainer.Terminate(ctx)
		require.NoError(t, err)
		cancel()
		<-ctx.Done()
	}
}

func runTests(ctx context.Context, t *testing.T, key *keys.PrivateKey, version string) {
	node := testContainerNode
	if version == localVersion {
		node = testLocalNode
	}

	cancel := runServer(ctx, t, node)
	defer cancel()

	var owner user.ID
	user.IDFromKey(&owner, key.PrivateKey.PublicKey)

	clientPool := getPool(ctx, t, key, node)
	cnrID := createContainer(ctx, t, clientPool, owner, containerName)
	restrictByEACL(ctx, t, clientPool, cnrID)

	t.Run("rest auth several tokens "+version, func(t *testing.T) { authTokens(ctx, t) })
	t.Run("rest check mix tokens up "+version, func(t *testing.T) { mixTokens(ctx, t, cnrID) })
	t.Run("rest form full binary bearer "+version, func(t *testing.T) { formFullBinaryBearer(ctx, t) })

	t.Run("rest put object "+version, func(t *testing.T) { restObjectPut(ctx, t, clientPool, cnrID) })
	t.Run("rest get object "+version, func(t *testing.T) { restObjectGet(ctx, t, clientPool, &owner, cnrID) })
	t.Run("rest get object unauthenticated "+version, func(t *testing.T) { restObjectGetUnauthenticated(ctx, t, clientPool, &owner, cnrID) })
	t.Run("rest get object full bearer "+version, func(t *testing.T) { restObjectGetFullBearer(ctx, t, clientPool, &owner, cnrID) })
	t.Run("rest delete object "+version, func(t *testing.T) { restObjectDelete(ctx, t, clientPool, &owner, cnrID) })
	t.Run("rest search objects "+version, func(t *testing.T) { restObjectsSearch(ctx, t, clientPool, &owner, cnrID) })

	t.Run("rest put container invalid "+version, func(t *testing.T) { restContainerPutInvalid(ctx, t) })
	t.Run("rest put container "+version, func(t *testing.T) { restContainerPut(ctx, t, clientPool) })
	t.Run("rest get container "+version, func(t *testing.T) { restContainerGet(ctx, t, owner, cnrID) })
	t.Run("rest delete container "+version, func(t *testing.T) { restContainerDelete(ctx, t, clientPool, owner) })
	t.Run("rest put container eacl "+version, func(t *testing.T) { restContainerEACLPut(ctx, t, clientPool, owner) })
	t.Run("rest get container eacl "+version, func(t *testing.T) { restContainerEACLGet(ctx, t, clientPool, cnrID) })
	t.Run("rest list containers	"+version, func(t *testing.T) { restContainerList(ctx, t, clientPool, owner, cnrID) })

	t.Run("rest manage storage group "+version, func(t *testing.T) { restManageStorageGroup(ctx, t, clientPool, owner, cnrID) })
}

func createDockerContainer(ctx context.Context, t *testing.T, image string) testcontainers.Container {
	req := testcontainers.ContainerRequest{
		Image:       image,
		WaitingFor:  wait.NewLogStrategy("aio container started").WithStartupTimeout(30 * time.Second),
		Name:        "aio",
		Hostname:    "aio",
		NetworkMode: "host",
	}
	aioC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	return aioC
}

func runServer(ctx context.Context, t *testing.T, node string) context.CancelFunc {
	cancelCtx, cancel := context.WithCancel(ctx)

	v := getDefaultConfig(node)
	l := newLogger(v)

	frostfsAPI, err := newFrostfsAPI(cancelCtx, l, v)
	require.NoError(t, err)

	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	require.NoError(t, err)

	api := operations.NewFrostfsRestGwAPI(swaggerSpec)
	server := restapi.NewServer(api, serverConfig(v))

	server.ConfigureAPI(frostfsAPI.Configure)

	go func() {
		err := server.Serve()
		require.NoError(t, err)
	}()

	return func() {
		cancel()
		err := server.Shutdown()
		require.NoError(t, err)
	}
}

func defaultHTTPClient() *http.Client {
	return &http.Client{Timeout: 60 * time.Second}
}

func getDefaultConfig(node string) *viper.Viper {
	v := config()
	v.SetDefault(cfgPeers+".0.address", node)
	v.SetDefault(cfgPeers+".0.weight", 1)
	v.SetDefault(cfgPeers+".0.priority", 1)
	v.SetDefault(cfgServerSection+restapi.FlagListenAddress, testListenAddress)
	v.SetDefault(cfgServerSection+restapi.FlagWriteTimeout, 60*time.Second)

	return v
}

func getPool(ctx context.Context, t *testing.T, key *keys.PrivateKey, node string) *pool.Pool {
	var prm pool.InitParameters
	prm.AddNode(pool.NewNodeParam(1, node, 1))
	prm.SetKey(&key.PrivateKey)
	prm.SetHealthcheckTimeout(5 * time.Second)
	prm.SetNodeDialTimeout(5 * time.Second)

	clientPool, err := pool.NewPool(prm)
	require.NoError(t, err)
	err = clientPool.Dial(ctx)
	require.NoError(t, err)

	return clientPool
}

func getRestrictBearerRecords() []*models.Record {
	return []*models.Record{
		formRestrictRecord(models.OperationGET),
		formRestrictRecord(models.OperationHEAD),
		formRestrictRecord(models.OperationPUT),
		formRestrictRecord(models.OperationDELETE),
		formRestrictRecord(models.OperationSEARCH),
		formRestrictRecord(models.OperationRANGE),
		formRestrictRecord(models.OperationRANGEHASH),
	}
}

func formRestrictRecord(op models.Operation) *models.Record {
	return &models.Record{
		Operation: models.NewOperation(op),
		Action:    models.NewAction(models.ActionDENY),
		Filters:   []*models.Filter{},
		Targets: []*models.Target{{
			Role: models.NewRole(models.RoleOTHERS),
			Keys: []string{},
		}}}
}

func authTokens(ctx context.Context, t *testing.T) {
	bearers := []*models.Bearer{
		{
			Name: "all-object",
			Object: []*models.Record{{
				Operation: models.NewOperation(models.OperationPUT),
				Action:    models.NewAction(models.ActionALLOW),
				Filters:   []*models.Filter{},
				Targets: []*models.Target{{
					Role: models.NewRole(models.RoleOTHERS),
					Keys: []string{},
				}},
			}},
		},
		{
			Name: "put-container",
			Container: &models.Rule{
				Verb: models.NewVerb(models.VerbPUT),
			},
		},
		{
			Name: "seteacl-container",
			Container: &models.Rule{
				Verb: models.NewVerb(models.VerbSETEACL),
			},
		},
		{
			Name: "delete-container",
			Container: &models.Rule{
				Verb: models.NewVerb(models.VerbDELETE),
			},
		},
	}

	httpClient := defaultHTTPClient()
	makeAuthTokenRequest(ctx, t, bearers, httpClient, false)
}

func mixTokens(ctx context.Context, t *testing.T, cnrID cid.ID) {
	bearers := []*models.Bearer{
		{
			Name: "all-object",
			Object: []*models.Record{{
				Operation: models.NewOperation(models.OperationPUT),
				Action:    models.NewAction(models.ActionALLOW),
				Filters:   []*models.Filter{},
				Targets: []*models.Target{{
					Role: models.NewRole(models.RoleOTHERS),
					Keys: []string{},
				}},
			}},
		},
		{
			Name: "put-container",
			Container: &models.Rule{
				Verb: models.NewVerb(models.VerbPUT),
			},
		},
		{
			Name: "seteacl-container",
			Container: &models.Rule{
				Verb: models.NewVerb(models.VerbSETEACL),
			},
		},
	}

	httpClient := defaultHTTPClient()
	tokens := makeAuthTokenRequest(ctx, t, bearers, httpClient, false)
	objectToken := tokens[0]
	containerPutToken := tokens[1]
	containerSetEACLToken := tokens[2]

	// check reject object token when container tokens is required
	checkPutContainerWithError(t, httpClient, objectToken)

	// check reject wrong verb container token
	checkPutContainerWithError(t, httpClient, containerSetEACLToken)

	// check reject wrong verb container token
	checkDeleteContainerWithError(t, httpClient, cnrID, containerSetEACLToken)

	// check reject wrong verb container token
	checkSetEACLContainerWithError(t, httpClient, cnrID, containerPutToken)

	// check reject container token when object tokens is required
	checkPutObjectWithError(t, httpClient, cnrID, containerSetEACLToken)
}

func formFullBinaryBearer(ctx context.Context, t *testing.T) {
	bearers := []*models.Bearer{
		{
			Name: "all-object",
			Object: []*models.Record{{
				Operation: models.NewOperation(models.OperationPUT),
				Action:    models.NewAction(models.ActionALLOW),
				Filters:   []*models.Filter{},
				Targets: []*models.Target{{
					Role: models.NewRole(models.RoleOTHERS),
					Keys: []string{},
				}},
			}},
		},
		{
			Name: "put-container",
			Container: &models.Rule{
				Verb: models.NewVerb(models.VerbPUT),
			},
		},
	}

	httpClient := defaultHTTPClient()
	tokens := makeAuthTokenRequest(ctx, t, bearers, httpClient, false)
	objectToken := tokens[0]
	containerPutToken := tokens[1]

	query := make(url.Values)
	query.Add(walletConnectQuery, strconv.FormatBool(useWalletConnect))

	// check that container token isn't valid
	request, err := http.NewRequest(http.MethodGet, testHost+"/v1/auth/bearer?"+query.Encode(), nil)
	require.NoError(t, err)
	prepareCommonHeaders(request.Header, containerPutToken)
	checkGWErrorResponse(t, httpClient, request)

	// check that object bearer token is valid
	request, err = http.NewRequest(http.MethodGet, testHost+"/v1/auth/bearer?"+query.Encode(), nil)
	require.NoError(t, err)
	prepareCommonHeaders(request.Header, objectToken)
	resp := &models.BinaryBearer{}
	doRequest(t, httpClient, request, http.StatusOK, resp)

	actualTokenRaw, err := base64.StdEncoding.DecodeString(*resp.Token)
	require.NoError(t, err)

	var actualToken bearer.Token
	err = actualToken.Unmarshal(actualTokenRaw)
	require.NoError(t, err)

	require.True(t, actualToken.VerifySignature())
	require.Len(t, actualToken.EACLTable().Records(), 1)
	actualRecord := actualToken.EACLTable().Records()[0]
	require.Equal(t, eacl.OperationPut, actualRecord.Operation())
	require.Equal(t, eacl.ActionAllow, actualRecord.Action())
	require.Empty(t, actualRecord.Filters())
	require.Len(t, actualRecord.Targets(), 1)
	actualTarget := actualRecord.Targets()[0]
	require.Empty(t, actualTarget.BinaryKeys())
	require.Equal(t, eacl.RoleOthers, actualTarget.Role())
}

func checkPutContainerWithError(t *testing.T, httpClient *http.Client, token *handlers.BearerToken) {
	reqURL, err := url.Parse(testHost + "/v1/containers")
	require.NoError(t, err)
	body, err := json.Marshal(&models.ContainerPutInfo{ContainerName: "container"})
	require.NoError(t, err)
	request, err := http.NewRequest(http.MethodPut, reqURL.String(), bytes.NewReader(body))
	require.NoError(t, err)
	prepareCommonHeaders(request.Header, token)

	checkGWErrorResponse(t, httpClient, request)
}

func checkDeleteContainerWithError(t *testing.T, httpClient *http.Client, cnrID cid.ID, token *handlers.BearerToken) {
	reqURL, err := url.Parse(testHost + "/v1/containers/" + cnrID.EncodeToString())
	require.NoError(t, err)
	request, err := http.NewRequest(http.MethodDelete, reqURL.String(), nil)
	require.NoError(t, err)
	prepareCommonHeaders(request.Header, token)

	checkGWErrorResponse(t, httpClient, request)
}

func checkSetEACLContainerWithError(t *testing.T, httpClient *http.Client, cnrID cid.ID, token *handlers.BearerToken) {
	req := models.Eacl{Records: []*models.Record{}}
	body, err := json.Marshal(&req)
	require.NoError(t, err)
	request, err := http.NewRequest(http.MethodPut, testHost+"/v1/containers/"+cnrID.EncodeToString()+"/eacl", bytes.NewReader(body))
	require.NoError(t, err)
	prepareCommonHeaders(request.Header, token)

	checkGWErrorResponse(t, httpClient, request)
}

func checkPutObjectWithError(t *testing.T, httpClient *http.Client, cnrID cid.ID, token *handlers.BearerToken) {
	req := &models.ObjectUpload{
		ContainerID: util.NewString(cnrID.EncodeToString()),
		FileName:    util.NewString("newFile.txt"),
		Payload:     base64.StdEncoding.EncodeToString([]byte("content")),
	}

	body, err := json.Marshal(req)
	require.NoError(t, err)

	request, err := http.NewRequest(http.MethodPut, testHost+"/v1/objects?", bytes.NewReader(body))
	require.NoError(t, err)
	prepareCommonHeaders(request.Header, token)

	checkGWErrorResponse(t, httpClient, request)
}

func checkGWErrorResponse(t *testing.T, httpClient *http.Client, request *http.Request) {
	resp := &models.ErrorResponse{}
	doRequest(t, httpClient, request, http.StatusBadRequest, resp)
	require.Equal(t, int64(0), resp.Code)
	require.Equal(t, models.ErrorTypeGW, *resp.Type)
}

func restManageStorageGroup(ctx context.Context, t *testing.T, clientPool *pool.Pool, owner user.ID, cnrID cid.ID) {
	attributes := map[string]string{object.AttributeFileName: "someFile"}
	content := []byte("content")
	objID := createObject(ctx, t, clientPool, &owner, cnrID, attributes, content)

	bearer := &models.Bearer{Object: getAllowedRules()}
	httpClient := defaultHTTPClient()
	bearerTokens := makeAuthTokenRequest(ctx, t, []*models.Bearer{bearer}, httpClient)
	bearerToken := bearerTokens[0]

	query := make(url.Values)
	query.Add(walletConnectQuery, strconv.FormatBool(useWalletConnect))

	var storageGroupAddress models.Address
	storageGroupName := "my-storage-group"
	t.Run("put storage group", func(t *testing.T) {
		req := &models.StorageGroupPutBody{
			Lifetime: util.NewInteger(10),
			Members:  []string{objID.EncodeToString()},
			Name:     storageGroupName,
		}

		body, err := json.Marshal(req)
		require.NoError(t, err)

		request, err := http.NewRequest(http.MethodPut, testHost+"/v1/containers/"+cnrID.EncodeToString()+"/storagegroups?"+query.Encode(), bytes.NewReader(body))
		require.NoError(t, err)
		prepareCommonHeaders(request.Header, bearerToken)

		doRequest(t, httpClient, request, http.StatusOK, &storageGroupAddress)
		require.Equal(t, cnrID.EncodeToString(), *storageGroupAddress.ContainerID)
	})

	t.Run("list storage groups", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodGet, testHost+"/v1/containers/"+cnrID.EncodeToString()+"/storagegroups?"+query.Encode(), nil)
		require.NoError(t, err)
		prepareCommonHeaders(request.Header, bearerToken)

		list := &models.StorageGroupList{}
		doRequest(t, httpClient, request, http.StatusOK, list)

		require.Equal(t, int64(1), *list.Size)
		require.Equal(t, storageGroupAddress.ContainerID, list.StorageGroups[0].Address.ContainerID)
		require.Equal(t, storageGroupAddress.ObjectID, list.StorageGroups[0].Address.ObjectID)
		require.Equal(t, storageGroupName, list.StorageGroups[0].Name)
		require.NotEmpty(t, list.StorageGroups[0].ExpirationEpoch)
	})

	t.Run("get storage group", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodGet, testHost+"/v1/containers/"+cnrID.EncodeToString()+"/storagegroups/"+*storageGroupAddress.ObjectID+"?"+query.Encode(), nil)
		require.NoError(t, err)
		prepareCommonHeaders(request.Header, bearerToken)

		group := &models.StorageGroup{}
		doRequest(t, httpClient, request, http.StatusOK, group)

		require.Equal(t, strconv.Itoa(len(content)), *group.Size)
		require.Equal(t, storageGroupName, group.Name)
		require.Equal(t, storageGroupAddress.ContainerID, group.Address.ContainerID)
		require.Equal(t, storageGroupAddress.ObjectID, group.Address.ObjectID)
		require.NotEmpty(t, *group.ExpirationEpoch)
		require.Equal(t, []string{objID.EncodeToString()}, group.Members)
	})

	t.Run("delete storage group", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodDelete, testHost+"/v1/containers/"+cnrID.EncodeToString()+"/storagegroups/"+*storageGroupAddress.ObjectID+"?"+query.Encode(), nil)
		require.NoError(t, err)
		prepareCommonHeaders(request.Header, bearerToken)

		resp := &models.SuccessResponse{}
		doRequest(t, httpClient, request, http.StatusOK, resp)
		require.True(t, *resp.Success)

		var sgObjID oid.ID
		err = sgObjID.DecodeString(*storageGroupAddress.ObjectID)
		require.NoError(t, err)

		var addr oid.Address
		addr.SetContainer(cnrID)
		addr.SetObject(sgObjID)

		var prm pool.PrmObjectHead
		prm.SetAddress(addr)

		_, err = clientPool.HeadObject(ctx, prm)
		require.Error(t, err)
	})
}

func getAllowedRules() []*models.Record {
	var result []*models.Record

	ops := []models.Operation{models.OperationGET, models.OperationHEAD, models.OperationPUT,
		models.OperationDELETE, models.OperationSEARCH, models.OperationRANGE, models.OperationRANGEHASH}

	for _, op := range ops {
		rec := &models.Record{
			Operation: models.NewOperation(op),
			Action:    models.NewAction(models.ActionALLOW),
			Filters:   []*models.Filter{},
			Targets: []*models.Target{{
				Role: models.NewRole(models.RoleOTHERS),
				Keys: []string{},
			}},
		}
		result = append(result, rec)
	}

	return result
}

func restObjectPut(ctx context.Context, t *testing.T, clientPool *pool.Pool, cnrID cid.ID) {
	bearer := &models.Bearer{
		Object: []*models.Record{{
			Operation: models.NewOperation(models.OperationPUT),
			Action:    models.NewAction(models.ActionALLOW),
			Filters:   []*models.Filter{},
			Targets: []*models.Target{{
				Role: models.NewRole(models.RoleOTHERS),
				Keys: []string{},
			}},
		}},
	}
	bearer.Object = append(bearer.Object, getRestrictBearerRecords()...)

	httpClient := defaultHTTPClient()
	bearerTokens := makeAuthTokenRequest(ctx, t, []*models.Bearer{bearer}, httpClient, false)
	bearerToken := bearerTokens[0]

	content := "content of file"
	attrKey, attrValue := "User-Attribute", "user value"

	attributes := map[string]string{
		object.AttributeFileName: "newFile.txt",
		attrKey:                  attrValue,
	}

	req := &models.ObjectUpload{
		ContainerID: util.NewString(cnrID.EncodeToString()),
		FileName:    util.NewString("newFile.txt"),
		Payload:     base64.StdEncoding.EncodeToString([]byte(content)),
		Attributes: []*models.Attribute{{
			Key:   &attrKey,
			Value: &attrValue,
		}},
	}

	body, err := json.Marshal(req)
	require.NoError(t, err)

	query := make(url.Values)
	query.Add(walletConnectQuery, strconv.FormatBool(useWalletConnect))

	request, err := http.NewRequest(http.MethodPut, testHost+"/v1/objects?"+query.Encode(), bytes.NewReader(body))
	require.NoError(t, err)
	prepareCommonHeaders(request.Header, bearerToken)

	addr := &models.Address{}
	doRequest(t, httpClient, request, http.StatusOK, addr)

	var CID cid.ID
	err = CID.DecodeString(*addr.ContainerID)
	require.NoError(t, err)
	var id oid.ID
	err = id.DecodeString(*addr.ObjectID)
	require.NoError(t, err)
	var objectAddress oid.Address
	objectAddress.SetContainer(CID)
	objectAddress.SetObject(id)

	var prm pool.PrmObjectGet
	prm.SetAddress(objectAddress)
	res, err := clientPool.GetObject(ctx, prm)
	require.NoError(t, err)

	payload := bytes.NewBuffer(nil)
	_, err = io.Copy(payload, res.Payload)
	require.NoError(t, err)
	require.Equal(t, content, payload.String())

	for _, attribute := range res.Header.Attributes() {
		require.Equal(t, attributes[attribute.Key()], attribute.Value(), attribute.Key())
	}
}

func restObjectGet(ctx context.Context, t *testing.T, p *pool.Pool, ownerID *user.ID, cnrID cid.ID) {
	content := []byte("some content")
	attributes := map[string]string{
		object.AttributeFileName: "get-obj-name",
		"user-attribute":         "user value",
	}

	objID := createObject(ctx, t, p, ownerID, cnrID, attributes, content)

	bearer := &models.Bearer{
		Object: []*models.Record{
			{
				Operation: models.NewOperation(models.OperationHEAD),
				Action:    models.NewAction(models.ActionALLOW),
				Filters:   []*models.Filter{},
				Targets: []*models.Target{{
					Role: models.NewRole(models.RoleOTHERS),
					Keys: []string{},
				}},
			},
			{
				Operation: models.NewOperation(models.OperationRANGE),
				Action:    models.NewAction(models.ActionALLOW),
				Filters:   []*models.Filter{},
				Targets: []*models.Target{{
					Role: models.NewRole(models.RoleOTHERS),
					Keys: []string{},
				}},
			},
		},
	}
	bearer.Object = append(bearer.Object, getRestrictBearerRecords()...)

	httpClient := defaultHTTPClient()
	bearerTokens := makeAuthTokenRequest(ctx, t, []*models.Bearer{bearer}, httpClient, false)
	bearerToken := bearerTokens[0]

	query := make(url.Values)
	query.Add(walletConnectQuery, strconv.FormatBool(useWalletConnect))

	request, err := http.NewRequest(http.MethodGet, testHost+"/v1/objects/"+cnrID.EncodeToString()+"/"+objID.EncodeToString()+"?"+query.Encode(), nil)
	require.NoError(t, err)
	prepareCommonHeaders(request.Header, bearerToken)

	objInfo := &models.ObjectInfo{}
	doRequest(t, httpClient, request, http.StatusOK, objInfo)

	require.Equal(t, cnrID.EncodeToString(), *objInfo.ContainerID)
	require.Equal(t, objID.EncodeToString(), *objInfo.ObjectID)
	require.Equal(t, ownerID.EncodeToString(), *objInfo.OwnerID)
	require.Equal(t, len(attributes), len(objInfo.Attributes))
	require.Equal(t, int64(len(content)), *objInfo.ObjectSize)

	contentData, err := base64.StdEncoding.DecodeString(objInfo.Payload)
	require.NoError(t, err)
	require.Equal(t, content, contentData)

	for _, attr := range objInfo.Attributes {
		require.Equal(t, attributes[*attr.Key], *attr.Value)
	}

	// check max-payload-size params
	query = make(url.Values)
	query.Add(walletConnectQuery, strconv.FormatBool(useWalletConnect))
	query.Add("max-payload-size", "0")

	request, err = http.NewRequest(http.MethodGet, testHost+"/v1/objects/"+cnrID.EncodeToString()+"/"+objID.EncodeToString()+"?"+query.Encode(), nil)
	require.NoError(t, err)
	prepareCommonHeaders(request.Header, bearerToken)

	objInfo = &models.ObjectInfo{}
	doRequest(t, httpClient, request, http.StatusOK, objInfo)
	require.Empty(t, objInfo.Payload)
	require.Equal(t, int64(0), *objInfo.PayloadSize)

	// check range params
	rangeLength := 4
	query = make(url.Values)
	query.Add(walletConnectQuery, strconv.FormatBool(useWalletConnect))
	query.Add("range-offset", "0")
	query.Add("range-length", strconv.Itoa(rangeLength))

	request, err = http.NewRequest(http.MethodGet, testHost+"/v1/objects/"+cnrID.EncodeToString()+"/"+objID.EncodeToString()+"?"+query.Encode(), nil)
	require.NoError(t, err)
	prepareCommonHeaders(request.Header, bearerToken)

	objInfo = &models.ObjectInfo{}
	doRequest(t, httpClient, request, http.StatusOK, objInfo)
	require.Equal(t, int64(rangeLength), *objInfo.PayloadSize)

	contentData, err = base64.StdEncoding.DecodeString(objInfo.Payload)
	require.NoError(t, err)
	require.Equal(t, content[:rangeLength], contentData)

	// check empty object
	objID2 := createObject(ctx, t, p, ownerID, cnrID, map[string]string{}, []byte{})

	query2 := make(url.Values)
	query2.Add(walletConnectQuery, strconv.FormatBool(useWalletConnect))

	request2, err := http.NewRequest(http.MethodGet, testHost+"/v1/objects/"+cnrID.EncodeToString()+"/"+objID2.EncodeToString()+"?"+query2.Encode(), nil)
	require.NoError(t, err)
	prepareCommonHeaders(request2.Header, bearerToken)

	objInfo2 := &models.ObjectInfo{}
	doRequest(t, httpClient, request2, http.StatusOK, objInfo2)

	require.Equal(t, cnrID.EncodeToString(), *objInfo2.ContainerID)
	require.Equal(t, objID2.EncodeToString(), *objInfo2.ObjectID)
	require.Equal(t, ownerID.EncodeToString(), *objInfo2.OwnerID)
	require.Equal(t, 0, len(objInfo2.Attributes))
	require.Equal(t, int64(0), *objInfo2.ObjectSize)
}

func restObjectGetUnauthenticated(ctx context.Context, t *testing.T, p *pool.Pool, ownerID *user.ID, cnrID cid.ID) {
	content := []byte("some content")
	attributes := map[string]string{
		object.AttributeFileName: "get-obj-unauth-name",
		"user-attribute":         "user value",
	}

	objID := createObject(ctx, t, p, ownerID, cnrID, attributes, content)

	httpClient := defaultHTTPClient()

	request, err := http.NewRequest(http.MethodGet, testHost+"/v1/objects/"+cnrID.EncodeToString()+"/"+objID.EncodeToString(), nil)
	require.NoError(t, err)

	request.Header.Add("Content-Type", "application/json")

	resp := &models.ErrorResponse{}
	doRequest(t, httpClient, request, http.StatusBadRequest, resp)
	require.Equal(t, int64(2048), resp.Code)
	require.Equal(t, models.ErrorTypeAPI, *resp.Type)

	// set empty eacl table to be able to do unauthenticated request
	allowByEACL(ctx, t, p, cnrID)

	request, err = http.NewRequest(http.MethodGet, testHost+"/v1/objects/"+cnrID.EncodeToString()+"/"+objID.EncodeToString(), nil)
	require.NoError(t, err)
	objInfo := &models.ObjectInfo{}
	doRequest(t, httpClient, request, http.StatusOK, objInfo)

	require.Equal(t, cnrID.EncodeToString(), *objInfo.ContainerID)
	require.Equal(t, objID.EncodeToString(), *objInfo.ObjectID)
	require.Equal(t, ownerID.EncodeToString(), *objInfo.OwnerID)
	require.Equal(t, len(attributes), len(objInfo.Attributes))
	require.Equal(t, int64(len(content)), *objInfo.ObjectSize)

	contentData, err := base64.StdEncoding.DecodeString(objInfo.Payload)
	require.NoError(t, err)
	require.Equal(t, content, contentData)

	for _, attr := range objInfo.Attributes {
		require.Equal(t, attributes[*attr.Key], *attr.Value)
	}

	// set eacl the same as was before test started
	restrictByEACL(ctx, t, p, cnrID)
}

func restObjectGetFullBearer(ctx context.Context, t *testing.T, p *pool.Pool, ownerID *user.ID, cnrID cid.ID) {
	content := []byte("some content")
	attributes := map[string]string{
		object.AttributeFileName: "get-obj-name",
		"user-attribute":         "user value",
	}

	objID := createObject(ctx, t, p, ownerID, cnrID, attributes, content)

	bearers := &models.Bearer{
		Object: []*models.Record{
			{
				Operation: models.NewOperation(models.OperationHEAD),
				Action:    models.NewAction(models.ActionALLOW),
				Filters:   []*models.Filter{},
				Targets: []*models.Target{{
					Role: models.NewRole(models.RoleOTHERS),
					Keys: []string{},
				}},
			},
			{
				Operation: models.NewOperation(models.OperationRANGE),
				Action:    models.NewAction(models.ActionALLOW),
				Filters:   []*models.Filter{},
				Targets: []*models.Target{{
					Role: models.NewRole(models.RoleOTHERS),
					Keys: []string{},
				}},
			},
		},
	}
	bearers.Object = append(bearers.Object, getRestrictBearerRecords()...)

	httpClient := defaultHTTPClient()
	bearerTokens := makeAuthTokenRequest(ctx, t, []*models.Bearer{bearers}, httpClient, true)
	bearerToken := bearerTokens[0]

	query := make(url.Values)
	query.Add(walletConnectQuery, strconv.FormatBool(useWalletConnect))

	request, err := http.NewRequest(http.MethodGet, testHost+"/v1/auth/bearer?"+query.Encode(), nil)
	require.NoError(t, err)
	prepareCommonHeaders(request.Header, bearerToken)
	resp := &models.BinaryBearer{}
	doRequest(t, httpClient, request, http.StatusOK, resp)
	actualTokenRaw, err := base64.StdEncoding.DecodeString(*resp.Token)
	require.NoError(t, err)
	var actualToken bearer.Token
	err = actualToken.Unmarshal(actualTokenRaw)
	require.NoError(t, err)
	// check that is token for all users
	require.True(t, actualToken.AssertUser(user.ID{}))

	query.Add(fullBearerQuery, "true")

	request, err = http.NewRequest(http.MethodGet, testHost+"/v1/objects/"+cnrID.EncodeToString()+"/"+objID.EncodeToString()+"?"+query.Encode(), nil)
	require.NoError(t, err)
	request.Header.Add("Authorization", "Bearer "+*resp.Token)

	objInfo := &models.ObjectInfo{}
	doRequest(t, httpClient, request, http.StatusOK, objInfo)
	contentData, err := base64.StdEncoding.DecodeString(objInfo.Payload)
	require.NoError(t, err)
	require.Equal(t, content, contentData)
}

func restObjectDelete(ctx context.Context, t *testing.T, p *pool.Pool, owner *user.ID, cnrID cid.ID) {
	objID := createObject(ctx, t, p, owner, cnrID, nil, []byte("some content"))

	bearer := &models.Bearer{
		Object: []*models.Record{{
			Operation: models.NewOperation(models.OperationDELETE),
			Action:    models.NewAction(models.ActionALLOW),
			Filters:   []*models.Filter{},
			Targets: []*models.Target{{
				Role: models.NewRole(models.RoleOTHERS),
				Keys: []string{},
			}},
		}},
	}
	bearer.Object = append(bearer.Object, getRestrictBearerRecords()...)

	httpClient := defaultHTTPClient()
	bearerTokens := makeAuthTokenRequest(ctx, t, []*models.Bearer{bearer}, httpClient, false)
	bearerToken := bearerTokens[0]

	query := make(url.Values)
	query.Add(walletConnectQuery, strconv.FormatBool(useWalletConnect))

	request, err := http.NewRequest(http.MethodDelete, testHost+"/v1/objects/"+cnrID.EncodeToString()+"/"+objID.EncodeToString()+"?"+query.Encode(), nil)
	require.NoError(t, err)
	prepareCommonHeaders(request.Header, bearerToken)

	resp := &models.SuccessResponse{}
	doRequest(t, httpClient, request, http.StatusOK, resp)
	require.True(t, *resp.Success)

	var addr oid.Address
	addr.SetContainer(cnrID)
	addr.SetObject(objID)

	var prm pool.PrmObjectHead
	prm.SetAddress(addr)

	_, err = p.HeadObject(ctx, prm)
	require.Error(t, err)
}

func restObjectsSearch(ctx context.Context, t *testing.T, p *pool.Pool, owner *user.ID, cnrID cid.ID) {
	userKey, userValue := "User-Attribute", "user-attribute-value"
	objectName := "object-name"
	filePath := "path/to/object/object-name"
	headers := map[string]string{
		object.AttributeFileName: objectName,
		"FilePath":               filePath,
		userKey:                  userValue,
	}
	objID := createObject(ctx, t, p, owner, cnrID, headers, []byte("some content"))
	headers[userKey] = "dummy"
	_ = createObject(ctx, t, p, owner, cnrID, headers, []byte("some content"))

	bearer := &models.Bearer{
		Object: []*models.Record{
			{
				Operation: models.NewOperation(models.OperationSEARCH),
				Action:    models.NewAction(models.ActionALLOW),
				Filters:   []*models.Filter{},
				Targets:   []*models.Target{{Role: models.NewRole(models.RoleOTHERS), Keys: []string{}}},
			},
			{
				Operation: models.NewOperation(models.OperationHEAD),
				Action:    models.NewAction(models.ActionALLOW),
				Filters:   []*models.Filter{},
				Targets:   []*models.Target{{Role: models.NewRole(models.RoleOTHERS), Keys: []string{}}},
			},
			{
				Operation: models.NewOperation(models.OperationGET),
				Action:    models.NewAction(models.ActionALLOW),
				Filters:   []*models.Filter{},
				Targets:   []*models.Target{{Role: models.NewRole(models.RoleOTHERS), Keys: []string{}}},
			},
		},
	}
	bearer.Object = append(bearer.Object, getRestrictBearerRecords()...)

	httpClient := defaultHTTPClient()
	bearerTokens := makeAuthTokenRequest(ctx, t, []*models.Bearer{bearer}, httpClient, false)
	bearerToken := bearerTokens[0]

	search := &models.SearchFilters{
		Filters: []*models.SearchFilter{
			{
				Key:   util.NewString(userKey),
				Match: models.NewSearchMatch(models.SearchMatchMatchStringEqual),
				Value: util.NewString(userValue),
			},
		},
	}

	body, err := json.Marshal(search)
	require.NoError(t, err)

	query := make(url.Values)
	query.Add(walletConnectQuery, strconv.FormatBool(useWalletConnect))

	request, err := http.NewRequest(http.MethodPost, testHost+"/v1/objects/"+cnrID.EncodeToString()+"/search?"+query.Encode(), bytes.NewReader(body))
	require.NoError(t, err)
	prepareCommonHeaders(request.Header, bearerToken)

	resp := &models.ObjectList{}
	doRequest(t, httpClient, request, http.StatusOK, resp)

	require.Equal(t, 1, int(*resp.Size))
	require.Len(t, resp.Objects, 1)

	objBaseInfo := resp.Objects[0]
	require.Equal(t, cnrID.EncodeToString(), *objBaseInfo.Address.ContainerID)
	require.Equal(t, objID.EncodeToString(), *objBaseInfo.Address.ObjectID)
	require.Equal(t, objectName, objBaseInfo.Name)
	require.Equal(t, filePath, objBaseInfo.FilePath)
}

func doRequest(t *testing.T, httpClient *http.Client, request *http.Request, expectedCode int, model interface{}) {
	resp, err := httpClient.Do(request)
	require.NoError(t, err)
	defer func() {
		err := resp.Body.Close()
		require.NoError(t, err)
	}()
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	if expectedCode != resp.StatusCode {
		fmt.Println("resp", string(respBody))
	}
	require.Equal(t, expectedCode, resp.StatusCode)

	if model == nil {
		return
	}

	err = json.Unmarshal(respBody, model)
	require.NoError(t, err)
}

func restContainerGet(ctx context.Context, t *testing.T, owner user.ID, cnrID cid.ID) {
	httpClient := &http.Client{Timeout: 5 * time.Second}
	request, err := http.NewRequest(http.MethodGet, testHost+"/v1/containers/"+cnrID.EncodeToString(), nil)
	require.NoError(t, err)
	request = request.WithContext(ctx)

	cnrInfo := &models.ContainerInfo{}
	doRequest(t, httpClient, request, http.StatusOK, cnrInfo)

	require.Equal(t, cnrID.EncodeToString(), *cnrInfo.ContainerID)
	require.Equal(t, owner.EncodeToString(), *cnrInfo.OwnerID)
	require.Equal(t, containerName, *cnrInfo.ContainerName)
	require.NotEmpty(t, *cnrInfo.Version)
}

func restContainerDelete(ctx context.Context, t *testing.T, clientPool *pool.Pool, owner user.ID) {
	cnrID := createContainer(ctx, t, clientPool, owner, "for-delete")

	bearer := &models.Bearer{
		Container: &models.Rule{
			Verb: models.NewVerb(models.VerbDELETE),
		},
	}

	httpClient := defaultHTTPClient()
	bearerTokens := makeAuthTokenRequest(ctx, t, []*models.Bearer{bearer}, httpClient, false)
	bearerToken := bearerTokens[0]

	query := make(url.Values)
	query.Add(walletConnectQuery, strconv.FormatBool(useWalletConnect))

	request, err := http.NewRequest(http.MethodDelete, testHost+"/v1/containers/"+cnrID.EncodeToString()+"?"+query.Encode(), nil)
	require.NoError(t, err)
	request = request.WithContext(ctx)
	prepareCommonHeaders(request.Header, bearerToken)

	resp := &models.SuccessResponse{}
	doRequest(t, httpClient, request, http.StatusOK, resp)
	require.True(t, *resp.Success)

	var prm pool.PrmContainerGet
	prm.SetContainerID(cnrID)

	_, err = clientPool.GetContainer(ctx, prm)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func restContainerEACLPut(ctx context.Context, t *testing.T, clientPool *pool.Pool, owner user.ID) {
	cnrID := createContainer(ctx, t, clientPool, owner, "for-eacl-put")
	httpClient := &http.Client{Timeout: 60 * time.Second}
	bearer := &models.Bearer{
		Container: &models.Rule{
			Verb: models.NewVerb(models.VerbSETEACL),
		},
	}
	bearerTokens := makeAuthTokenRequest(ctx, t, []*models.Bearer{bearer}, httpClient, false)
	bearerToken := bearerTokens[0]

	req := models.Eacl{
		Records: []*models.Record{{
			Action:    models.NewAction(models.ActionDENY),
			Filters:   []*models.Filter{},
			Operation: models.NewOperation(models.OperationDELETE),
			Targets: []*models.Target{{
				Keys: []string{"031a6c6fbbdf02ca351745fa86b9ba5a9452d785ac4f7fc2b7548ca2a46c4fcf4a"},
				Role: models.NewRole(models.RoleOTHERS),
			}},
		}},
	}

	invalidBody, err := json.Marshal(&req)
	require.NoError(t, err)

	req.Records[0].Targets[0].Role = models.NewRole(models.RoleKEYS)
	body, err := json.Marshal(&req)
	require.NoError(t, err)

	query := make(url.Values)
	query.Add(walletConnectQuery, strconv.FormatBool(useWalletConnect))

	doSetEACLRequest(ctx, t, httpClient, cnrID, query, bearerToken, invalidBody, http.StatusBadRequest, nil)

	resp := &models.SuccessResponse{}
	doSetEACLRequest(ctx, t, httpClient, cnrID, query, bearerToken, body, http.StatusOK, resp)
	require.True(t, *resp.Success)

	var prm pool.PrmContainerEACL
	prm.SetContainerID(cnrID)

	table, err := clientPool.GetEACL(ctx, prm)
	require.NoError(t, err)

	expectedTable, err := util.ToNativeTable(req.Records)
	require.NoError(t, err)
	expectedTable.SetCID(cnrID)

	require.True(t, eacl.EqualTables(*expectedTable, table))
}

func doSetEACLRequest(ctx context.Context, t *testing.T, httpClient *http.Client, cnrID cid.ID, query url.Values, bearerToken *handlers.BearerToken, body []byte, status int, model interface{}) {
	request, err := http.NewRequest(http.MethodPut, testHost+"/v1/containers/"+cnrID.EncodeToString()+"/eacl?"+query.Encode(), bytes.NewReader(body))
	require.NoError(t, err)
	request = request.WithContext(ctx)
	prepareCommonHeaders(request.Header, bearerToken)

	doRequest(t, httpClient, request, status, model)
}

func restContainerEACLGet(ctx context.Context, t *testing.T, p *pool.Pool, cnrID cid.ID) {
	var prm pool.PrmContainerEACL
	prm.SetContainerID(cnrID)
	expectedTable, err := p.GetEACL(ctx, prm)
	require.NoError(t, err)

	httpClient := &http.Client{Timeout: 60 * time.Second}

	request, err := http.NewRequest(http.MethodGet, testHost+"/v1/containers/"+cnrID.EncodeToString()+"/eacl", nil)
	require.NoError(t, err)
	request = request.WithContext(ctx)

	responseTable := &models.Eacl{}
	doRequest(t, httpClient, request, http.StatusOK, responseTable)

	require.Equal(t, cnrID.EncodeToString(), responseTable.ContainerID)

	actualTable, err := util.ToNativeTable(responseTable.Records)
	require.NoError(t, err)
	actualTable.SetCID(cnrID)

	require.True(t, eacl.EqualTables(expectedTable, *actualTable))
}

func restContainerList(ctx context.Context, t *testing.T, p *pool.Pool, owner user.ID, cnrID cid.ID) {
	var prm pool.PrmContainerList
	prm.SetOwnerID(owner)

	ids, err := p.ListContainers(ctx, prm)
	require.NoError(t, err)

	httpClient := defaultHTTPClient()

	query := make(url.Values)
	query.Add("ownerId", owner.EncodeToString())

	request, err := http.NewRequest(http.MethodGet, testHost+"/v1/containers?"+query.Encode(), nil)
	require.NoError(t, err)
	request = request.WithContext(ctx)

	list := &models.ContainerList{}
	doRequest(t, httpClient, request, http.StatusOK, list)

	require.Equal(t, len(ids), int(*list.Size))

	require.Truef(t, containsContainer(list.Containers, cnrID.EncodeToString(), containerName), "list doesn't contain cnr '%s' with name '%s'", cnrID.EncodeToString(), containerName)
}

func containsContainer(containers []*models.ContainerInfo, cnrID, cnrName string) bool {
	for _, cnrInfo := range containers {
		if *cnrInfo.ContainerID == cnrID {
			for _, attr := range cnrInfo.Attributes {
				if *attr.Key == "Name" && *attr.Value == cnrName {
					return true
				}
			}

			fmt.Println("container found but name doesn't match")
			return false
		}
	}

	return false
}

func makeAuthTokenRequest(ctx context.Context, t *testing.T, bearers []*models.Bearer, httpClient *http.Client, forAllUsers bool) []*handlers.BearerToken {
	key, err := keys.NewPrivateKeyFromHex(devenvPrivateKey)
	require.NoError(t, err)

	var ownerID user.ID
	user.IDFromKey(&ownerID, key.PrivateKey.PublicKey)

	data, err := json.Marshal(bearers)
	require.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, testHost+"/v1/auth", bytes.NewReader(data))
	require.NoError(t, err)
	request = request.WithContext(ctx)
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add(XBearerOwnerID, ownerID.String())
	request.Header.Add(XBearerForAllUsers, strconv.FormatBool(forAllUsers))

	resp, err := httpClient.Do(request)
	require.NoError(t, err)
	defer func() {
		err := resp.Body.Close()
		require.NoError(t, err)
	}()

	rr, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	if resp.StatusCode != http.StatusOK {
		fmt.Println("auth response", string(rr))
	}
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var stokenResp []*models.TokenResponse
	err = json.Unmarshal(rr, &stokenResp)
	require.NoError(t, err)

	fmt.Println("resp tokens:")

	respTokens := make([]*handlers.BearerToken, len(stokenResp))
	for i, tok := range stokenResp {
		isObject, err := handlers.IsObjectToken(bearers[i])
		require.NoError(t, err)

		require.Equal(t, bearers[i].Name, tok.Name)

		if isObject {
			require.Equal(t, models.TokenTypeObject, *tok.Type)
		} else {
			require.Equal(t, models.TokenTypeContainer, *tok.Type)
		}

		binaryData, err := base64.StdEncoding.DecodeString(*tok.Token)
		require.NoError(t, err)

		var bt *handlers.BearerToken
		if useWalletConnect {
			bt = signTokenWalletConnect(t, key, binaryData)
		} else {
			bt = signToken(t, key, binaryData)
		}

		respTokens[i] = bt
		fmt.Printf("%+v\n", bt)
	}

	return respTokens
}

func signToken(t *testing.T, key *keys.PrivateKey, data []byte) *handlers.BearerToken {
	signer := frostfsecdsa.Signer(key.PrivateKey)
	sign, err := signer.Sign(data)
	require.NoError(t, err)

	return &handlers.BearerToken{
		Token:     base64.StdEncoding.EncodeToString(data),
		Signature: hex.EncodeToString(sign),
		Key:       hex.EncodeToString(key.PublicKey().Bytes()),
	}
}

func signTokenWalletConnect(t *testing.T, key *keys.PrivateKey, data []byte) *handlers.BearerToken {
	signer := frostfsecdsa.SignerWalletConnect(key.PrivateKey)
	signature, err := signer.Sign(data)
	require.NoError(t, err)

	return &handlers.BearerToken{
		Token:     base64.StdEncoding.EncodeToString(data),
		Signature: hex.EncodeToString(signature),
		Key:       hex.EncodeToString(key.PublicKey().Bytes()),
	}
}

func restContainerPutInvalid(ctx context.Context, t *testing.T) {
	bearer := &models.Bearer{
		Container: &models.Rule{
			Verb: models.NewVerb(models.VerbPUT),
		},
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	bearerTokens := makeAuthTokenRequest(ctx, t, []*models.Bearer{bearer}, httpClient, false)
	bearerToken := bearerTokens[0]

	reqURL, err := url.Parse(testHost + "/v1/containers")
	require.NoError(t, err)
	query := reqURL.Query()
	query.Add("name-scope-global", "true")
	query.Add(walletConnectQuery, strconv.FormatBool(useWalletConnect))
	reqURL.RawQuery = query.Encode()

	body, err := json.Marshal(&models.ContainerPutInfo{ContainerName: "nameWithCapitalLetters"})
	require.NoError(t, err)
	request, err := http.NewRequest(http.MethodPut, reqURL.String(), bytes.NewReader(body))
	require.NoError(t, err)
	prepareCommonHeaders(request.Header, bearerToken)

	resp := &models.ErrorResponse{}
	doRequest(t, httpClient, request, http.StatusBadRequest, resp)
	require.Equal(t, int64(0), resp.Code)
	require.Equal(t, models.ErrorTypeGW, *resp.Type)
}

func restContainerPut(ctx context.Context, t *testing.T, clientPool *pool.Pool) {
	bearer := &models.Bearer{
		Container: &models.Rule{
			Verb: models.NewVerb(models.VerbPUT),
		},
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}
	bearerTokens := makeAuthTokenRequest(ctx, t, []*models.Bearer{bearer}, httpClient, false)
	bearerToken := bearerTokens[0]

	attrKey, attrValue := "User-Attribute", "user value"
	userAttributes := map[string]string{
		attrKey: attrValue,
	}

	// try to create container without name but with name-scope-global
	body, err := json.Marshal(&models.ContainerPutInfo{})
	require.NoError(t, err)

	reqURL, err := url.Parse(testHost + "/v1/containers")
	require.NoError(t, err)
	query := reqURL.Query()
	query.Add("name-scope-global", "true")
	query.Add(walletConnectQuery, strconv.FormatBool(useWalletConnect))
	reqURL.RawQuery = query.Encode()

	request, err := http.NewRequest(http.MethodPut, reqURL.String(), bytes.NewReader(body))
	require.NoError(t, err)
	prepareCommonHeaders(request.Header, bearerToken)

	doRequest(t, httpClient, request, http.StatusBadRequest, nil)

	// create container with name in local scope
	containerPutInfo := &models.ContainerPutInfo{
		Attributes: []*models.Attribute{{
			Key:   util.NewString(attrKey),
			Value: util.NewString(attrValue),
		}},
	}
	body, err = json.Marshal(containerPutInfo)
	require.NoError(t, err)

	reqURL, err = url.Parse(testHost + "/v1/containers")
	require.NoError(t, err)
	query = reqURL.Query()
	query.Add(walletConnectQuery, strconv.FormatBool(useWalletConnect))
	reqURL.RawQuery = query.Encode()

	request, err = http.NewRequest(http.MethodPut, reqURL.String(), bytes.NewReader(body))
	require.NoError(t, err)
	prepareCommonHeaders(request.Header, bearerToken)

	addr := &operations.PutContainerOKBody{}
	doRequest(t, httpClient, request, http.StatusOK, addr)

	var CID cid.ID
	err = CID.DecodeString(*addr.ContainerID)
	require.NoError(t, err)
	fmt.Println(CID.String())

	var prm pool.PrmContainerGet
	prm.SetContainerID(CID)

	cnr, err := clientPool.GetContainer(ctx, prm)
	require.NoError(t, err)

	cnrAttr := make(map[string]string)
	cnr.IterateAttributes(func(key, val string) {
		cnrAttr[key] = val
	})

	for key, val := range userAttributes {
		require.Equal(t, val, cnrAttr[key])
	}
}

func prepareCommonHeaders(header http.Header, bearerToken *handlers.BearerToken) {
	header.Add("Content-Type", "application/json")
	header.Add(XBearerSignature, bearerToken.Signature)
	header.Add("Authorization", "Bearer "+bearerToken.Token)
	header.Add(XBearerSignatureKey, bearerToken.Key)
}

func createContainer(ctx context.Context, t *testing.T, clientPool *pool.Pool, owner user.ID, name string) cid.ID {
	var policy netmap.PlacementPolicy
	err := policy.DecodeString("REP 1")
	require.NoError(t, err)

	var cnr container.Container
	cnr.Init()
	cnr.SetOwner(owner)
	cnr.SetPlacementPolicy(policy)
	cnr.SetBasicACL(acl.PublicRWExtended)

	container.SetName(&cnr, name)
	container.SetCreationTime(&cnr, time.Now())

	err = pool.SyncContainerWithNetwork(ctx, &cnr, clientPool)
	require.NoError(t, err)

	var waitPrm pool.WaitParams
	waitPrm.SetPollInterval(3 * time.Second)
	waitPrm.SetTimeout(30 * time.Second)

	var prm pool.PrmContainerPut
	prm.SetContainer(cnr)
	prm.SetWaitParams(waitPrm)

	CID, err := clientPool.PutContainer(ctx, prm)
	require.NoError(t, err)

	return CID
}

func createObject(ctx context.Context, t *testing.T, p *pool.Pool, ownerID *user.ID, cnrID cid.ID, headers map[string]string, payload []byte) oid.ID {
	attributes := make([]object.Attribute, 0, len(headers))

	for key, val := range headers {
		attr := object.NewAttribute()
		attr.SetKey(key)
		attr.SetValue(val)
		attributes = append(attributes, *attr)
	}

	obj := object.New()
	obj.SetOwnerID(ownerID)
	obj.SetContainerID(cnrID)
	obj.SetAttributes(attributes...)
	obj.SetPayload(payload)

	var prm pool.PrmObjectPut
	prm.SetHeader(*obj)

	objID, err := p.PutObject(ctx, prm)
	require.NoError(t, err)

	return objID
}

func restrictByEACL(ctx context.Context, t *testing.T, clientPool *pool.Pool, cnrID cid.ID) *eacl.Table {
	table := eacl.NewTable()
	table.SetCID(cnrID)

	for op := eacl.OperationGet; op <= eacl.OperationRangeHash; op++ {
		record := new(eacl.Record)
		record.SetOperation(op)
		record.SetAction(eacl.ActionDeny)
		target := new(eacl.Target)
		target.SetRole(eacl.RoleOthers)
		record.SetTargets(*target)
		table.AddRecord(record)
	}

	var waitPrm pool.WaitParams
	waitPrm.SetPollInterval(3 * time.Second)
	waitPrm.SetTimeout(30 * time.Second)

	var prm pool.PrmContainerSetEACL
	prm.SetTable(*table)
	prm.SetWaitParams(waitPrm)

	err := clientPool.SetEACL(ctx, prm)
	require.NoError(t, err)

	return table
}

func allowByEACL(ctx context.Context, t *testing.T, clientPool *pool.Pool, cnrID cid.ID) *eacl.Table {
	table := eacl.NewTable()
	table.SetCID(cnrID)

	var waitPrm pool.WaitParams
	waitPrm.SetPollInterval(3 * time.Second)
	waitPrm.SetTimeout(15 * time.Second)

	var prm pool.PrmContainerSetEACL
	prm.SetTable(*table)
	prm.SetWaitParams(waitPrm)

	err := clientPool.SetEACL(ctx, prm)
	require.NoError(t, err)

	return table
}
