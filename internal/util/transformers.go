package util

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/TrueCloudLab/frostfs-rest-gw/gen/models"
	"github.com/TrueCloudLab/frostfs-sdk-go/bearer"
	apistatus "github.com/TrueCloudLab/frostfs-sdk-go/client/status"
	cid "github.com/TrueCloudLab/frostfs-sdk-go/container/id"
	"github.com/TrueCloudLab/frostfs-sdk-go/eacl"
	"github.com/TrueCloudLab/frostfs-sdk-go/object"
	"github.com/TrueCloudLab/frostfs-sdk-go/session"
)

// ToNativeAction converts models.Action to appropriate eacl.Action.
func ToNativeAction(a *models.Action) (eacl.Action, error) {
	if a == nil {
		return eacl.ActionUnknown, fmt.Errorf("unsupported empty action")
	}

	switch *a {
	case models.ActionALLOW:
		return eacl.ActionAllow, nil
	case models.ActionDENY:
		return eacl.ActionDeny, nil
	default:
		return eacl.ActionUnknown, fmt.Errorf("unsupported action type: '%s'", *a)
	}
}

// FromNativeAction converts eacl.Action to appropriate models.Action.
func FromNativeAction(a eacl.Action) (*models.Action, error) {
	switch a {
	case eacl.ActionAllow:
		return models.NewAction(models.ActionALLOW), nil
	case eacl.ActionDeny:
		return models.NewAction(models.ActionDENY), nil
	default:
		return nil, fmt.Errorf("unsupported action type: '%s'", a)
	}
}

// ToNativeOperation converts models.Operation to appropriate eacl.Operation.
func ToNativeOperation(o *models.Operation) (eacl.Operation, error) {
	if o == nil {
		return eacl.OperationUnknown, fmt.Errorf("unsupported empty opertaion")
	}

	switch *o {
	case models.OperationGET:
		return eacl.OperationGet, nil
	case models.OperationHEAD:
		return eacl.OperationHead, nil
	case models.OperationPUT:
		return eacl.OperationPut, nil
	case models.OperationDELETE:
		return eacl.OperationDelete, nil
	case models.OperationSEARCH:
		return eacl.OperationSearch, nil
	case models.OperationRANGE:
		return eacl.OperationRange, nil
	case models.OperationRANGEHASH:
		return eacl.OperationRangeHash, nil
	default:
		return eacl.OperationUnknown, fmt.Errorf("unsupported operation type: '%s'", *o)
	}
}

// FromNativeOperation converts eacl.Operation to appropriate models.Operation.
func FromNativeOperation(o eacl.Operation) (*models.Operation, error) {
	switch o {
	case eacl.OperationGet:
		return models.NewOperation(models.OperationGET), nil
	case eacl.OperationHead:
		return models.NewOperation(models.OperationHEAD), nil
	case eacl.OperationPut:
		return models.NewOperation(models.OperationPUT), nil
	case eacl.OperationDelete:
		return models.NewOperation(models.OperationDELETE), nil
	case eacl.OperationSearch:
		return models.NewOperation(models.OperationSEARCH), nil
	case eacl.OperationRange:
		return models.NewOperation(models.OperationRANGE), nil
	case eacl.OperationRangeHash:
		return models.NewOperation(models.OperationRANGEHASH), nil
	default:
		return nil, fmt.Errorf("unsupported operation type: '%s'", o)
	}
}

// ToNativeHeaderType converts models.HeaderType to appropriate eacl.FilterHeaderType.
func ToNativeHeaderType(h *models.HeaderType) (eacl.FilterHeaderType, error) {
	if h == nil {
		return eacl.HeaderTypeUnknown, fmt.Errorf("unsupported empty header type")
	}

	switch *h {
	case models.HeaderTypeOBJECT:
		return eacl.HeaderFromObject, nil
	case models.HeaderTypeREQUEST:
		return eacl.HeaderFromRequest, nil
	case models.HeaderTypeSERVICE:
		return eacl.HeaderFromService, nil
	default:
		return eacl.HeaderTypeUnknown, fmt.Errorf("unsupported header type: '%s'", *h)
	}
}

// FromNativeHeaderType converts eacl.FilterHeaderType to appropriate models.HeaderType.
func FromNativeHeaderType(h eacl.FilterHeaderType) (*models.HeaderType, error) {
	switch h {
	case eacl.HeaderFromObject:
		return models.NewHeaderType(models.HeaderTypeOBJECT), nil
	case eacl.HeaderFromRequest:
		return models.NewHeaderType(models.HeaderTypeREQUEST), nil
	case eacl.HeaderFromService:
		return models.NewHeaderType(models.HeaderTypeSERVICE), nil
	default:
		return nil, fmt.Errorf("unsupported header type: '%s'", h)
	}
}

// ToNativeMatchType converts models.MatchType to appropriate eacl.Match.
func ToNativeMatchType(t *models.MatchType) (eacl.Match, error) {
	if t == nil {
		return eacl.MatchUnknown, fmt.Errorf("unsupported empty match type")
	}

	switch *t {
	case models.MatchTypeSTRINGEQUAL:
		return eacl.MatchStringEqual, nil
	case models.MatchTypeSTRINGNOTEQUAL:
		return eacl.MatchStringNotEqual, nil
	default:
		return eacl.MatchUnknown, fmt.Errorf("unsupported match type: '%s'", *t)
	}
}

// FromNativeMatchType converts eacl.Match to appropriate models.MatchType.
func FromNativeMatchType(t eacl.Match) (*models.MatchType, error) {
	switch t {
	case eacl.MatchStringEqual:
		return models.NewMatchType(models.MatchTypeSTRINGEQUAL), nil
	case eacl.MatchStringNotEqual:
		return models.NewMatchType(models.MatchTypeSTRINGNOTEQUAL), nil
	default:
		return nil, fmt.Errorf("unsupported match type: '%s'", t)
	}
}

// ToNativeRole converts models.Role to appropriate eacl.Role.
func ToNativeRole(r *models.Role) (eacl.Role, error) {
	if r == nil {
		return eacl.RoleUnknown, fmt.Errorf("unsupported empty role")
	}

	switch *r {
	case models.RoleUSER:
		return eacl.RoleUser, nil
	case models.RoleSYSTEM:
		return eacl.RoleSystem, nil
	case models.RoleOTHERS:
		return eacl.RoleOthers, nil
	case models.RoleKEYS:
		return eacl.RoleUnknown, nil
	default:
		return 0, fmt.Errorf("unsupported role type: '%s'", *r)
	}
}

// FromNativeRole converts eacl.Role to appropriate models.Role.
func FromNativeRole(r eacl.Role) (*models.Role, error) {
	switch r {
	case eacl.RoleUser:
		return models.NewRole(models.RoleUSER), nil
	case eacl.RoleSystem:
		return models.NewRole(models.RoleSYSTEM), nil
	case eacl.RoleOthers:
		return models.NewRole(models.RoleOTHERS), nil
	case eacl.RoleUnknown:
		return models.NewRole(models.RoleKEYS), nil
	default:
		return nil, fmt.Errorf("unsupported role type: '%s'", r)
	}
}

// ToNativeVerb converts models.Verb to appropriate session.ContainerSessionVerb.
func ToNativeVerb(r *models.Verb) (session.ContainerVerb, error) {
	if r == nil {
		return 0, fmt.Errorf("unsupported empty verb type")
	}

	switch *r {
	case models.VerbPUT:
		return session.VerbContainerPut, nil
	case models.VerbDELETE:
		return session.VerbContainerDelete, nil
	case models.VerbSETEACL:
		return session.VerbContainerSetEACL, nil
	default:
		return 0, fmt.Errorf("unsupported verb type: '%s'", *r)
	}
}

// ToNativeContainerToken converts models.Rule to appropriate session.Token.
func ToNativeContainerToken(tokenRule *models.Rule) (session.Container, error) {
	var tok session.Container

	if tokenRule.ContainerID != "" {
		var cnrID cid.ID
		if err := cnrID.DecodeString(tokenRule.ContainerID); err != nil {
			return session.Container{}, fmt.Errorf("couldn't parse container id: %w", err)
		}
		tok.ApplyOnlyTo(cnrID)
	}

	verb, err := ToNativeVerb(tokenRule.Verb)
	if err != nil {
		return session.Container{}, err
	}
	tok.ForVerb(verb)

	return tok, nil
}

// ToNativeRecord converts models.Record to appropriate eacl.Record.
func ToNativeRecord(r *models.Record) (*eacl.Record, error) {
	var record eacl.Record

	action, err := ToNativeAction(r.Action)
	if err != nil {
		return nil, err
	}
	record.SetAction(action)

	operation, err := ToNativeOperation(r.Operation)
	if err != nil {
		return nil, err
	}
	record.SetOperation(operation)

	for _, filter := range r.Filters {
		headerType, err := ToNativeHeaderType(filter.HeaderType)
		if err != nil {
			return nil, err
		}
		matchType, err := ToNativeMatchType(filter.MatchType)
		if err != nil {
			return nil, err
		}
		if filter.Key == nil || filter.Value == nil {
			return nil, fmt.Errorf("invalid filter")
		}
		record.AddFilter(headerType, matchType, *filter.Key, *filter.Value)
	}

	targets := make([]eacl.Target, len(r.Targets))
	for i, target := range r.Targets {
		trgt, err := ToNativeTarget(target)
		if err != nil {
			return nil, err
		}
		targets[i] = *trgt
	}
	record.SetTargets(targets...)

	return &record, nil
}

// FromNativeRecord converts eacl.Record to appropriate models.Record.
func FromNativeRecord(r eacl.Record) (*models.Record, error) {
	var err error
	var record models.Record

	record.Action, err = FromNativeAction(r.Action())
	if err != nil {
		return nil, err
	}

	record.Operation, err = FromNativeOperation(r.Operation())
	if err != nil {
		return nil, err
	}

	record.Filters = make([]*models.Filter, len(r.Filters()))
	for i, filter := range r.Filters() {
		headerType, err := FromNativeHeaderType(filter.From())
		if err != nil {
			return nil, err
		}
		matchType, err := FromNativeMatchType(filter.Matcher())
		if err != nil {
			return nil, err
		}

		record.Filters[i] = &models.Filter{
			HeaderType: headerType,
			Key:        NewString(filter.Key()),
			MatchType:  matchType,
			Value:      NewString(filter.Value()),
		}
	}

	record.Targets = make([]*models.Target, len(r.Targets()))
	for i, target := range r.Targets() {
		trgt, err := FromNativeTarget(target)
		if err != nil {
			return nil, err
		}
		record.Targets[i] = trgt
	}

	return &record, nil
}

// ToNativeTarget converts models.Target to appropriate eacl.Target.
func ToNativeTarget(t *models.Target) (*eacl.Target, error) {
	var target eacl.Target

	if len(t.Keys) > 0 && *t.Role != models.RoleKEYS {
		return nil, fmt.Errorf("you cannot set binary keys with role other than '%s'", models.RoleKEYS)
	}

	role, err := ToNativeRole(t.Role)
	if err != nil {
		return nil, err
	}
	target.SetRole(role)

	keys := make([][]byte, len(t.Keys))
	for i, key := range t.Keys {
		binaryKey, err := hex.DecodeString(key)
		if err != nil {
			return nil, fmt.Errorf("couldn't decode target key: %w", err)
		}
		keys[i] = binaryKey
	}
	target.SetBinaryKeys(keys)

	return &target, nil
}

// FromNativeTarget converts eacl.Target to appropriate models.Target.
func FromNativeTarget(t eacl.Target) (*models.Target, error) {
	var err error
	var target models.Target

	target.Role, err = FromNativeRole(t.Role())
	if err != nil {
		return nil, err
	}

	target.Keys = make([]string, len(t.BinaryKeys()))
	for i, key := range t.BinaryKeys() {
		target.Keys[i] = hex.EncodeToString(key)
	}

	return &target, nil
}

// ToNativeObjectToken converts []*models.Record to appropriate token.BearerToken.
func ToNativeObjectToken(tokenRecords []*models.Record) (*bearer.Token, error) {
	table, err := ToNativeTable(tokenRecords)
	if err != nil {
		return nil, err
	}

	var btoken bearer.Token
	btoken.SetEACLTable(*table)

	return &btoken, nil
}

// ToNativeTable converts records to eacl.Table.
func ToNativeTable(records []*models.Record) (*eacl.Table, error) {
	table := eacl.NewTable()

	for _, rec := range records {
		record, err := ToNativeRecord(rec)
		if err != nil {
			return nil, fmt.Errorf("couldn't transform record to native: %w", err)
		}
		table.AddRecord(record)
	}

	return table, nil
}

// ToNativeMatchFilter converts models.SearchMatch to object.SearchMatchType.
func ToNativeMatchFilter(s *models.SearchMatch) (object.SearchMatchType, error) {
	if s == nil {
		return object.MatchUnknown, fmt.Errorf("unsupported empty verb type")
	}

	switch *s {
	case models.SearchMatchMatchStringEqual:
		return object.MatchStringEqual, nil
	case models.SearchMatchMatchStringNotEqual:
		return object.MatchStringNotEqual, nil
	case models.SearchMatchMatchNotPresent:
		return object.MatchNotPresent, nil
	case models.SearchMatchMatchCommonPrefix:
		return object.MatchCommonPrefix, nil
	default:
		return object.MatchUnknown, fmt.Errorf("unsupported search match: '%s'", *s)
	}
}

// ToNativeFilters converts models.SearchFilters to object.SearchFilters.
func ToNativeFilters(fs *models.SearchFilters) (object.SearchFilters, error) {
	filters := object.NewSearchFilters()
	filters.AddRootFilter()

	for _, f := range fs.Filters {
		matchFilter, err := ToNativeMatchFilter(f.Match)
		if err != nil {
			return nil, err
		}

		filters.AddFilter(*f.Key, *f.Value, matchFilter)
	}

	return filters, nil
}

// NewString returns pointer to provided string.
func NewString(val string) *string {
	return &val
}

// NewInteger returns pointer to provided int.
func NewInteger(val int64) *int64 {
	return &val
}

// NewBool returns pointer to provided bool.
func NewBool(val bool) *bool {
	return &val
}

// NewSuccessResponse forms model.SuccessResponse.
func NewSuccessResponse() *models.SuccessResponse {
	return &models.SuccessResponse{
		Success: NewBool(true),
	}
}

// NewErrorResponse forms model.ErrorResponse.
func NewErrorResponse(err error) *models.ErrorResponse {
	var code int64
	t := models.ErrorTypeGW
	if status, ok := unwrapErr(err).(apistatus.StatusV2); ok {
		code = int64(status.ToStatusV2().Code())
		t = models.ErrorTypeAPI
	}

	return &models.ErrorResponse{
		Code:    code,
		Message: NewString(err.Error()),
		Type:    models.NewErrorType(t),
	}
}

func unwrapErr(err error) error {
	for e := errors.Unwrap(err); e != nil; e = errors.Unwrap(err) {
		err = e
	}
	return err
}
