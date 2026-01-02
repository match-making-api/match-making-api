package common

// GetReadableFields retorna os campos básicos de leitura + campos adicionais
func GetReadableFields(additionalFields ...map[string]bool) map[string]bool {
	baseFields := map[string]bool{
		"ID":            ALLOW,
		"RIDSource":     ALLOW,
		"SourceKey":     ALLOW,
		"ResourceOwner": ALLOW,
		"CreatedAt":     ALLOW,
		"UpdatedAt":     ALLOW,
	}

	return MergeFieldMaps(baseFields, additionalFields...)
}

// GetQueryableFields retorna os campos básicos de consulta + campos adicionais
func GetQueryableFields(additionalFields ...map[string]bool) map[string]bool {
	baseFields := map[string]bool{
		"ID":        ALLOW,
		"CreatedAt": ALLOW,
		"UpdatedAt": ALLOW,
	}

	return MergeFieldMaps(baseFields, additionalFields...)
}

// MergeFieldMaps combina múltiplos mapas de campos
func MergeFieldMaps(baseMap map[string]bool, additionalMaps ...map[string]bool) map[string]bool {
	result := make(map[string]bool)

	// Adiciona o mapa base
	for k, v := range baseMap {
		result[k] = v
	}

	// Adiciona os mapas adicionais
	for _, m := range additionalMaps {
		for k, v := range m {
			result[k] = v
		}
	}

	return result
}
