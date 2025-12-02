package common

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/google/uuid"
)

type QueryService[T any] interface {
	GetName() string
	GetByID(ctx context.Context, id uuid.UUID) (*T, error)
	Search(ctx context.Context, s Search) ([]*T, error)
	Compile(ctx context.Context, searchParams []SearchAggregation, resultOptions SearchResultOptions) (*Search, error)
}

type BaseQueryService[T any, R Searchable[T]] struct {
	Repository      R
	QueryableFields map[string]bool
	ReadableFields  map[string]bool
	MaxPageSize     uint
	Audience        IntendedAudienceKey
	name            string
}

func (service *BaseQueryService[T, R]) GetName() string {
	if service.name != "" {
		return service.name
	}

	service.name = reflect.TypeOf(service).Name()

	return service.name
}

// / GetByID returns a single entity by its ID using ClientApplicationAudienceIDKey as the intended audience.
func (service *BaseQueryService[T, R]) GetByID(ctx context.Context, id uuid.UUID) (*T, error) {
	entity, err := service.Repository.GetByID(ctx, id)

	if err != nil {
		var typeDef T
		typeName := reflect.TypeOf(typeDef).Name()
		svcName := service.GetName()
		return nil, fmt.Errorf("error searching. Service: %v. Entity: %v. Error: %v", svcName, typeName, err)
	}

	return entity, nil
}

func (service *BaseQueryService[T, R]) Search(ctx context.Context, s Search) ([]*T, error) {
	var omitFields []string
	var pickFields []string
	for fieldName, isReadable := range service.ReadableFields {
		if !isReadable {
			omitFields = append(omitFields, fieldName)
			continue
		}

		pickFields = append(pickFields, fieldName)
	}

	if len(omitFields) > 0 {
		slog.Info("Omitting fields", "fields", omitFields)
	}

	s.ResultOptions.OmitFields = omitFields
	s.ResultOptions.PickFields = pickFields

	entities, err := service.Repository.Search(ctx, s)

	if err != nil {
		var typeDef T
		typeName := reflect.TypeOf(typeDef).Name()
		svcName := service.GetName()
		return nil, fmt.Errorf("error filtering. Service: %v. Entity: %v. Error: %v", svcName, typeName, err)
	}

	return entities, nil
}

func (svc *BaseQueryService[T, R]) Compile(ctx context.Context, searchParams []SearchAggregation, resultOptions SearchResultOptions) (*Search, error) {
	err := ValidateSearchParameters(searchParams, svc.QueryableFields)
	if err != nil {
		return nil, fmt.Errorf("error validating search parameters: %v", err)
	}

	err = ValidateResultOptions(resultOptions, svc.ReadableFields)
	if err != nil {
		return nil, fmt.Errorf("error validating result options: %v", err)
	}

	intendedAud := GetIntendedAudience(ctx)

	if intendedAud == nil {
		intendedAud = &svc.Audience
	}

	s := NewSearchByAggregation(ctx, searchParams, resultOptions, *intendedAud)

	return &s, nil
}
