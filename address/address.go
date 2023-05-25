package address

import (
	"context"
	"fmt"

	"authservice/builder"
	"authservice/helper"
	"authservice/models"
	"authservice/repository"
)

type Address interface {
	GetAddresses(ctx context.Context, userId string) ([]*models.Address, error)
	GetAddress(ctx context.Context, userId, id string) (*models.Address, error)
	CreateAddress(ctx context.Context, userId string, address *models.Address) (*models.Address, error)
	UpdateAddress(ctx context.Context, userId, id string, address *models.Address) (*models.Address, error)
	DeleteAddress(ctx context.Context, userId, id string) error
}

type address struct {
	helper helper.Helper
	postgres repository.PostgresQueryer
	builder  builder.AddressBuilder
}

func NewAddress(b builder.AddressBuilder, h helper.Helper, p repository.PostgresQueryer) Address {
	return &address{
		postgres: p,
		builder: b,
		helper: h,
	}
}

func (a *address) GetAddresses(ctx context.Context, userId string) ([]*models.Address, error) {
	query := a.builder.GetAddresses()
	res, err := a.postgres.QueryScan(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("GetAddresses: unable to execute query: %s", err)
	}

	result := make([]*models.Address, 0)
	for res.Next() {
		var addr models.Address
		err = res.Scan(&addr)
		if err != nil {
			return nil, fmt.Errorf("GetAddresses: unable to parse result: %s", err)
		}

		result = append(result, &addr)
	}

	return result, nil
}

func (a *address) GetAddress(ctx context.Context, userId, id string) (*models.Address, error) {
	query := a.builder.GetAddress()
	res, err := a.postgres.QueryScan(ctx, query, userId, id)
	if err != nil {
		return nil, fmt.Errorf("GetAddress: unable to execute query: %s", err)
	}

	var addr models.Address
	if res.Next() {
		err = res.Scan(&addr)
		if err != nil {
			return nil, fmt.Errorf("GetAddress: unable to parse result: %s", err)
		}
	}

	return &addr, nil
}

func (a *address) CreateAddress(ctx context.Context, userId string, address *models.Address) (*models.Address, error) {
	id := a.helper.NewId()
	address.Id = &id
	address.UserId = &userId
	_ = a.removeDefaults(ctx, userId, address.IsDefault)
	query := a.builder.CreateAddress(address.GetFieldMap())
	_, err := a.postgres.Exec(ctx, query, userId, id)
	if err != nil {
		return nil, fmt.Errorf("CreateAddress: unable to execute query: %s", err)
	}

	return address, nil
}

func (a *address) removeDefaults(ctx context.Context, userId string, isDefault *bool) error {
	if isDefault != nil {
		if *isDefault {
			query := a.builder.RemoveDefaults()
			_, err := a.postgres.Exec(ctx, query, userId)

			return err
		}
	}

	return nil
}

func (a *address) UpdateAddress(ctx context.Context, userId, id string, address *models.Address) (*models.Address, error) {
	_ = a.removeDefaults(ctx, userId, address.IsDefault)
	query := a.builder.UpdateAddress(address.GetFieldMap())
	res, err := a.postgres.Exec(ctx, query, userId, id)
	if err != nil {
		return nil, fmt.Errorf("UpdateAddress: unable to execute query: %s", err)
	}

	if res == 0 {
		return nil, fmt.Errorf("UpdateAddress: no such address to update")
	}

	address.Id = &id
	return address, nil
}

func (a *address) DeleteAddress(ctx context.Context, userId, id string) error {
	query := a.builder.DeleteAddress()
	res, err := a.postgres.Exec(ctx, query, userId, id)
	if err != nil {
		return fmt.Errorf("DeleteAddress: unable to execute query: %s", err)
	}

	if res == 0 {
		return fmt.Errorf("DeleteAddress: no such address to delete")
	}

	return nil
}
