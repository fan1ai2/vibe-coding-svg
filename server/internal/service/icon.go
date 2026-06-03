package service

import (
	"github.com/fan1ai2/vibe-coding-svg/server/internal/model"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/neo4j"
	"github.com/fan1ai2/vibe-coding-svg/server/internal/repo"
)

type IconService struct {
	iconRepo    *repo.IconRepo
	tagRepo     *repo.TagRepo
	graphSync   *neo4j.GraphSyncService
}

func NewIconService(iconRepo *repo.IconRepo, tagRepo *repo.TagRepo, graphSync *neo4j.GraphSyncService) *IconService {
	return &IconService{iconRepo: iconRepo, tagRepo: tagRepo, graphSync: graphSync}
}

type CreateIconInput struct {
	Name       string
	SvgContent string
	IsPublic   bool
	Tags       []TagInput
	Theme      string
}

type TagInput struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (s *IconService) Create(userID string, input CreateIconInput) (*model.Icon, error) {
	icon, err := s.iconRepo.Create(userID, input.Name, input.SvgContent, input.IsPublic)
	if err != nil {
		return nil, err
	}

	// Extract colors from SVG
	colors := neo4j.ExtractColors(input.SvgContent, 5)

	// Attach tags
	tagModels := make([]model.Tag, 0)
	neo4jTags := make([]neo4j.TagData, 0)
	for _, ti := range input.Tags {
		tagType := ti.Type
		if tagType == "" {
			tagType = "usage"
		}
		tag, err := s.tagRepo.FindOrCreate(ti.Name, tagType)
		if err != nil {
			return nil, err
		}
		tagModels = append(tagModels, *tag)
		neo4jTags = append(neo4jTags, neo4j.TagData{Name: tag.Name, Slug: tag.Slug, Type: tag.Type})
	}

	tagIDs := make([]string, len(tagModels))
	for i, t := range tagModels {
		tagIDs[i] = t.ID
	}

	if err := s.iconRepo.AttachTags(icon.ID, tagIDs); err != nil {
		return nil, err
	}
	if err := s.iconRepo.AttachColors(icon.ID, colors); err != nil {
		return nil, err
	}
	if err := s.iconRepo.AttachTheme(icon.ID, input.Theme); err != nil {
		return nil, err
	}

	// Async Neo4j sync
	s.graphSync.SyncCreate(icon.ID, icon.Name, neo4jTags, colors, input.Theme)

	icon.Tags = tagModels
	icon.Colors = colors
	icon.Theme = input.Theme
	return icon, nil
}

type BatchIconInput struct {
	Name       string     `json:"name"`
	SvgContent string     `json:"svg_content"`
	IsPublic   bool       `json:"is_public"`
	Tags       []TagInput `json:"tags"`
	Theme      string     `json:"theme"`
}

func (s *IconService) BatchCreate(userID string, inputs []BatchIconInput) ([]*model.Icon, error) {
	// Batch PG insert
	pgInputs := make([]repo.BatchIcon, len(inputs))
	for i, in := range inputs {
		pgInputs[i].Name = in.Name
		pgInputs[i].SvgContent = in.SvgContent
		pgInputs[i].IsPublic = in.IsPublic
		pgInputs[i].Theme = in.Theme
	}

	icons, err := s.iconRepo.BatchCreate(userID, pgInputs)
	if err != nil {
		return nil, err
	}

	// Attach tags, colors, and sync Neo4j for each icon
	for i, icon := range icons {
		colors := neo4j.ExtractColors(inputs[i].SvgContent, 5)
		_ = s.iconRepo.AttachColors(icon.ID, colors)

		neo4jTags := make([]neo4j.TagData, 0)
		tagIDs := make([]string, 0)
		for _, ti := range inputs[i].Tags {
			tagType := ti.Type
			if tagType == "" {
				tagType = "usage"
			}
			tag, err := s.tagRepo.FindOrCreate(ti.Name, tagType)
			if err != nil {
				continue
			}
			tagIDs = append(tagIDs, tag.ID)
			neo4jTags = append(neo4jTags, neo4j.TagData{Name: tag.Name, Slug: tag.Slug, Type: tag.Type})
		}
		_ = s.iconRepo.AttachTags(icon.ID, tagIDs)
		if inputs[i].Theme != "" {
			_ = s.iconRepo.AttachTheme(icon.ID, inputs[i].Theme)
		}

		s.graphSync.SyncCreate(icon.ID, icon.Name, neo4jTags, colors, inputs[i].Theme)
	}

	return icons, nil
}

func (s *IconService) GetByID(id string) (*model.Icon, error) {
	icon, err := s.iconRepo.FindByID(id)
	if err != nil || icon == nil {
		return icon, err
	}
	tags, _ := s.iconRepo.LoadTags(id)
	colors, _ := s.iconRepo.LoadColors(id)
	icon.Tags = tags
	icon.Colors = colors
	return icon, nil
}

func (s *IconService) ListPublic(limit, offset int) ([]*model.Icon, error) {
	return s.iconRepo.ListPublic(limit, offset)
}

func (s *IconService) ListByUser(userID string, limit, offset int) ([]*model.Icon, error) {
	return s.iconRepo.FindByUserID(userID, limit, offset)
}

func (s *IconService) Delete(iconID, userID string) error {
	icon, err := s.iconRepo.FindByID(iconID)
	if err != nil {
		return err
	}
	if icon == nil {
		return nil
	}
	if icon.UserID != userID {
		return nil // handled by handler
	}
	if err := s.iconRepo.Delete(iconID); err != nil {
		return err
	}
	s.graphSync.SyncDelete(iconID)
	return nil
}

func (s *IconService) Search(params repo.IconSearchParams) ([]*model.Icon, error) {
	return s.iconRepo.Search(params)
}

func (s *IconService) Recommend(iconID string, limit int) ([]*model.Icon, error) {
	recommend, err := neo4j.GetRelatedIcons(iconID, limit)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(recommend))
	for i, r := range recommend {
		ids[i] = r.IconID
	}
	return s.iconRepo.FindByIDs(ids)
}
