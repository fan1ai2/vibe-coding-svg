package neo4j

import (
	"fmt"
	"log"
)

type IconNode struct {
	ID   string
	Name string
}

type RecommendResult struct {
	IconID string
	Name   string
	Score  int64
}

// CreateIconNode creates or updates an Icon node in Neo4j.
func CreateIconNode(icon IconNode) error {
	s := NewSession()
	defer s.Close(driverCtx())
	_, err := s.Run(driverCtx(),
		`MERGE (i:Icon {id: $id}) SET i.name = $name`,
		map[string]interface{}{"id": icon.ID, "name": icon.Name},
	)
	return err
}

// CreateIconRelations creates Tag/Color/Theme nodes and relationships for an Icon.
func CreateIconRelations(iconID string, tags []TagData, colors []string, theme string) error {
	s := NewSession()
	defer s.Close(driverCtx())

	for _, tag := range tags {
		_, err := s.Run(driverCtx(),
			`MERGE (t:Tag {slug: $slug})
			 SET t.name = $name, t.type = $type
			 WITH t
			 MATCH (i:Icon {id: $iconID})
			 MERGE (i)-[:HAS_TAG]->(t)`,
			map[string]interface{}{
				"slug":    tag.Slug,
				"name":   tag.Name,
				"type":   tag.Type,
				"iconID": iconID,
			},
		)
		if err != nil {
			return fmt.Errorf("tag %s: %w", tag.Name, err)
		}
	}

	for _, color := range colors {
		_, err := s.Run(driverCtx(),
			`MERGE (c:Color {hex: $hex})
			 WITH c
			 MATCH (i:Icon {id: $iconID})
			 MERGE (i)-[:HAS_COLOR]->(c)`,
			map[string]interface{}{"hex": color, "iconID": iconID},
		)
		if err != nil {
			return fmt.Errorf("color %s: %w", color, err)
		}
	}

	if theme != "" {
		_, err := s.Run(driverCtx(),
			`MERGE (t:Theme {name: $name})
			 WITH t
			 MATCH (i:Icon {id: $iconID})
			 MERGE (i)-[:IN_THEME]->(t)`,
			map[string]interface{}{"name": theme, "iconID": iconID},
		)
		if err != nil {
			return fmt.Errorf("theme %s: %w", theme, err)
		}
	}

	return nil
}

type TagData struct {
	Name string
	Slug string
	Type string
}

// DeleteIconNode removes an Icon node and its relationships from Neo4j.
func DeleteIconNode(iconID string) error {
	s := NewSession()
	defer s.Close(driverCtx())
	_, err := s.Run(driverCtx(),
		`MATCH (i:Icon {id: $id}) DETACH DELETE i`,
		map[string]interface{}{"id": iconID},
	)
	return err
}

// GetRelatedIcons finds icons sharing tags, colors, or themes with the given icon.
// Uses tag type weighting: usage×3, style×2, category×1.
func GetRelatedIcons(iconID string, limit int) ([]RecommendResult, error) {
	if limit <= 0 {
		limit = 10
	}

	s := NewReadSession()
	defer s.Close(driverCtx())

	result, err := s.Run(driverCtx(),
		`MATCH (i:Icon {id: $id})-[r:HAS_TAG|HAS_COLOR|IN_THEME]->(attr)
		      <-[:HAS_TAG|HAS_COLOR|IN_THEME]-(related:Icon)
		 WHERE related.id <> i.id
		 WITH related, r, attr,
		      CASE type(r)
		        WHEN 'HAS_TAG' THEN
		          CASE attr.type WHEN 'usage' THEN 3 WHEN 'style' THEN 2 ELSE 1 END
		        ELSE 1
		      END AS weight
		 RETURN related.id AS id, related.name AS name, sum(weight) AS score
		 ORDER BY score DESC LIMIT $limit`,
		map[string]interface{}{"id": iconID, "limit": limit},
	)
	if err != nil {
		return nil, fmt.Errorf("recommend query: %w", err)
	}

	list := make([]RecommendResult, 0)
	for result.Next(driverCtx()) {
		record := result.Record()
		id, _ := record.Get("id")
		name, _ := record.Get("name")
		score, _ := record.Get("score")
		list = append(list, RecommendResult{
			IconID: fmt.Sprint(id),
			Name:   fmt.Sprint(name),
			Score:  score.(int64),
		})
	}
	if err := result.Err(); err != nil {
		return nil, err
	}
	return list, result.Err()
}

// GraphSyncService handles async Neo4j sync operations.
type GraphSyncService struct{}

func NewGraphSyncService() *GraphSyncService {
	return &GraphSyncService{}
}

// SyncCreate asynchronously creates Neo4j nodes and relations for a new icon.
func (svc *GraphSyncService) SyncCreate(iconID, iconName string, tags []TagData, colors []string, theme string) {
	go func() {
		if err := CreateIconNode(IconNode{ID: iconID, Name: iconName}); err != nil {
			log.Printf("[neo4j] sync create node %s: %v", iconID, err)
			return
		}
		if err := CreateIconRelations(iconID, tags, colors, theme); err != nil {
			log.Printf("[neo4j] sync create relations %s: %v", iconID, err)
			return
		}
	}()
}

// SyncDelete asynchronously removes Neo4j node for a deleted icon.
func (svc *GraphSyncService) SyncDelete(iconID string) {
	go func() {
		if err := DeleteIconNode(iconID); err != nil {
			log.Printf("[neo4j] sync delete %s: %v", iconID, err)
		}
	}()
}
