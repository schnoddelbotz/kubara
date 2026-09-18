package catalog

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"

	jsonpatch "github.com/evanphx/json-patch/v5"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"
)

type LoadOptions struct {
	CWD              string
	BootstrapCatalog string
	Catalogs         []string
	Overwrite        bool
	MergeServices    bool
}

func Load(options LoadOptions) (Catalog, error) {
	merged := Catalog{
		Services: make(map[string]ServiceDefinition),
	}

	sources, err := ResolveSources(options)
	if err != nil {
		return Catalog{}, err
	}

	bootstrap, err := loadCatalogSource(sources[0])
	if err != nil {
		return Catalog{}, fmt.Errorf("load bootstrap catalog: %w", err)
	}

	for name, def := range bootstrap.Services {
		merged.Services[name] = def
	}

	for _, cat := range sources[1:] {
		external, err := loadCatalogSource(cat)
		if err != nil {
			return Catalog{}, fmt.Errorf("load catalog %q: %w", cat, err)
		}

		for name, def := range external.Services {
			// if options.MergeServices {
			_, exists := merged.Services[name]
			merged.Services[name], err = mergeServiceDefinitions(exists, merged.Services[name], def)
			if err != nil {
				return merged, err
			}
			// continue
			// }
			// if _, exists := merged.Services[name]; exists && !options.Overwrite {
			// 	return Catalog{}, fmt.Errorf("service definition %q already exists in another catalog", name)
			// }
			// merged.Services[name] = def
		}
	}

	return merged, err
}

func mergeServiceDefinitions(exists bool, dst, src ServiceDefinition) (ServiceDefinition, error) {
	var err error
	// dst does not exist = no merge, no overwrite
	if !exists {
		return src, err
	}
	// dst has nil cs, use src cs
	if dst.Spec.ConfigSchema == nil && src.Spec.ConfigSchema != nil {
		dst.Spec.ConfigSchema = src.Spec.ConfigSchema
		return dst, err
	}
	// dst and src have cs, must merge
	dst.Spec.ConfigSchema, err = mergeJSONSchemaProps(dst.Spec.ConfigSchema, src.Spec.ConfigSchema)
	return dst, err
}

func mergeJSONSchemaProps(dst, src *apiextensionsv1.JSONSchemaProps) (*apiextensionsv1.JSONSchemaProps, error) {
	if dst == nil && src == nil {
		return nil, nil
	}
	if dst == nil {
		return src.DeepCopy(), nil
	}
	if src == nil {
		return dst.DeepCopy(), nil
	}

	dstBytes, err := json.Marshal(dst)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dst schema: %w", err)
	}

	srcBytes, err := json.Marshal(src)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal src schema: %w", err)
	}

	mergedBytes, err := jsonpatch.MergePatch(dstBytes, srcBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to merge json patch: %w", err)
	}

	var result apiextensionsv1.JSONSchemaProps
	if err := json.Unmarshal(mergedBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal merged schema: %w", err)
	}

	return &result, nil
}

func loadCatalogSource(reference string) (Catalog, error) {
	source, err := ResolveSource(reference)
	if err != nil {
		return Catalog{}, fmt.Errorf("resolve catalog source: %w", err)
	}

	loaded, err := loadFromFS(os.DirFS(source.ServicesPath), ".")
	if err != nil {
		return Catalog{}, fmt.Errorf("load catalog from %q: %w", source.ServicesPath, err)
	}

	return loaded, nil
}

func loadFromFS(fsys fs.FS, root string) (Catalog, error) {
	catalog := Catalog{Services: map[string]ServiceDefinition{}}

	var files []string
	if err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		lowerPath := strings.ToLower(path)
		if !strings.HasSuffix(lowerPath, ".yaml") && !strings.HasSuffix(lowerPath, ".yml") {
			return nil
		}
		files = append(files, path)
		return nil
	}); err != nil {
		return Catalog{}, fmt.Errorf("walk service definitions: %w", err)
	}

	sort.Strings(files)
	for _, path := range files {
		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return Catalog{}, fmt.Errorf("read %q: %w", path, err)
		}

		var definition ServiceDefinition
		if err := yaml.Unmarshal(content, &definition); err != nil {
			return Catalog{}, fmt.Errorf("unmarshal %q: %w", path, err)
		}
		if err := definition.Validate(); err != nil {
			return Catalog{}, fmt.Errorf("invalid service definition %q: %w", path, err)
		}

		if _, exists := catalog.Services[definition.Metadata.Name]; exists {
			return Catalog{}, fmt.Errorf("duplicate service definition %q in %q", definition.Metadata.Name, path)
		}
		catalog.Services[definition.Metadata.Name] = definition
	}

	return catalog, nil
}
