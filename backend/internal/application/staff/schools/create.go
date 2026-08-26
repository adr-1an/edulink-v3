package schools

import (
	"app/internal/helpers/staff"
	"app/internal/helpers/staff/schools"
	"context"
	"log"
	"strings"

	"golang.org/x/text/language"
)

type CreateSchool struct {
	Name       string
	RegionCode string
}

type CreateSchoolInput struct {
	UserID int64
	School CreateSchool
}

func (s *Service) CreateSchool(ctx context.Context, i *CreateSchoolInput) error {
	// Parse & validate
	i.School.Name = strings.TrimSpace(i.School.Name)
	i.School.RegionCode = strings.TrimSpace(strings.ToUpper(strings.ReplaceAll(i.School.RegionCode, " ", "")))

	if i.School.Name == "" || len(i.School.Name) > 64 {
		return ErrInvalidName
	}

	if i.School.RegionCode != "" {
		region, err := language.ParseRegion(i.School.RegionCode)
		if err != nil {
			return ErrBadRequest
		}

		if !region.IsCountry() {
			return ErrInvalidRegionCode
		}
	}

	// Generate ID
	id, err := s.Sf.NextID()
	if err != nil {
		log.Println(err)
		return err
	}

	// Start tx
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Store
	res, err := tx.ExecContext(ctx, `
		INSERT INTO schools (id, owner_id, name, region_code)
		VALUES ($1, $2, $3, $4)
	`, id, i.UserID, i.School.Name, i.School.RegionCode)
	if err != nil {
		log.Println(err)
		return err
	}

	if err := staff_helpers.RowsAffected(res); err != nil {
		return err
	}

	// Log action
	if err := schools.StoreSchoolLog(id, i.UserID, schools.ActionSchoolCreate, schools.TypeCreate, "School created", "{user} created the school '"+i.School.Name+"'.", tx, ctx, s.Sf, "{user} created school '"+i.School.Name+"' in region '"+i.School.RegionCode+"'."); err != nil {
		log.Println(err)
		return err
	}

	// Commit tx
	if err := tx.Commit(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}
