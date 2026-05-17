package capabilities

func cloneManifest(in Manifest) Manifest {
	out := in
	out.ToolRefs = append([]string(nil), in.ToolRefs...)
	out.DefaultAllowedRoles = append([]string(nil), in.DefaultAllowedRoles...)
	out.AllowedRoles = append([]string(nil), in.AllowedRoles...)
	if in.Metadata != nil {
		out.Metadata = make(map[string]any, len(in.Metadata))
		for key, value := range in.Metadata {
			out.Metadata[key] = value
		}
	}
	return out
}
