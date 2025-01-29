// sdkTargetsFromResourceTargets converts the given Kubernetes resource targets to AWS service targets
func sdkTargetsFromResourceTargets(
	targets []*svcapitypes.Target,
) ([]*svcsdktypes.Target, error) {
	var res []*svcsdktypes.Target
	{{- $field := (index .CRD.SpecFields "Targets" )}}
	for _, krTarget := range targets {
		t := &svcsdktypes.Target{}
		{{ GoCodeSetSDKForStruct .CRD "" "t" $field.ShapeRef.Shape.MemberRef "" "krTarget" 1 }}
		res = append(res, t)
	}
	return res, nil
}

// resourceTargetsFromSDKTargets converts the given AWS service targets to Kubernetes resource targets
func resourceTargetsFromSDKTargets(
	targets []*svcsdktypes.Target,
) ([]*svcapitypes.Target) {
	var res []*svcapitypes.Target
	for _, sdkTarget := range targets {
		t := &svcapitypes.Target{}
		// test
		{{- $operation := (index .SDKAPI.API.Operations "PutTargets")}}
		{{- $targetsSDKShape := (index $operation.InputRef.Shape.MemberRefs "Targets")}} 
		{{ GoCodeSetResourceForStruct .CRD "" "t"  $field.ShapeRef.Shape.MemberRef "sdkTarget" $targetsSDKShape.Shape.MemberRef 1 }}
		res = append(res, t)
	}
	return res
}
