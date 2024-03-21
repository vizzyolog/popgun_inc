package schemaReg

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

const schemaDir = "schemaReg/schemas"

// SchemaInfo хранит информацию о схеме, включая саму схему и ее версию.
type SchemaInfo struct {
	Schema  *gojsonschema.Schema
	Version int
}

// SchemaRegistry теперь хранит схемы в виде вложенных карт,
// где внешний ключ - это название сущности, внутренний ключ - тип действия,
// а значение - список схем с информацией о версии.
type SchemaRegistry struct {
	schemas map[string]map[string][]SchemaInfo
}

func NewSchemaRegistry() *SchemaRegistry {
	wd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("Не удалось получить рабочую директорию: %v", err))
	}
	fmt.Printf("Текущая рабочая директория: %s\n", wd)
	schemas, err := loadSchemas(schemaDir)
	if err != nil {
		panic(err)
	}
	return &SchemaRegistry{
		schemas: schemas,
	}
}

func (sr *SchemaRegistry) Validate(entity, action string, version int, document interface{}) error {
	actionSchemas, ok := sr.schemas[entity][action]
	if !ok {
		return fmt.Errorf("schemas for entity %s and action %s not found", entity, action)
	}

	for _, schemaInfo := range actionSchemas {
		if schemaInfo.Version == version {
			result, err := schemaInfo.Schema.Validate(gojsonschema.NewGoLoader(document))
			if err != nil {
				return err
			}
			if !result.Valid() {
				return fmt.Errorf("document is not valid: %s", result.Errors())
			}
			return nil
		}
	}

	return fmt.Errorf("version %v not found for entity %s and action %s", version, entity, action)
}

func loadSchemas(dirPath string) (map[string]map[string][]SchemaInfo, error) {
	schemas := make(map[string]map[string][]SchemaInfo)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			relPath, _ := filepath.Rel(dirPath, path)
			parts := strings.Split(relPath, string(os.PathSeparator))
			if len(parts) != 3 { // Ожидаем структуру: сущность/действие/версия.json
				return fmt.Errorf("unexpected file structure: %s", relPath)
			}
			entity, action, versionFile := parts[0], parts[1], parts[2]
			version := strings.TrimSuffix(versionFile, filepath.Ext(versionFile))
			versionInt, err := strconv.Atoi(version)
			if err != nil {
				return err
			}
			schemaLoader := gojsonschema.NewReferenceLoader("file://" + path)
			schema, err := gojsonschema.NewSchema(schemaLoader)
			if err != nil {
				return err
			}

			if _, ok := schemas[entity]; !ok {
				schemas[entity] = make(map[string][]SchemaInfo)
			}
			schemas[entity][action] = append(schemas[entity][action], SchemaInfo{Schema: schema, Version: versionInt})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return schemas, nil
}
