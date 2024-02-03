// SdkTargetsFromResourceTargets converts the given Kubernetes resource targets to AWS service targets
func SdkTargetsFromResourceTargets(
	targets []*svcapitypes.Target,
) ([]*svcsdk.Target) {
	var res []*svcsdk.Target
	{{- $field := (index .CRD.SpecFields "Targets" )}}
	for _, krTarget := range targets {
		t := &svcsdk.Target{}
		{{ GoCodeSetSDKForStruct .CRD "" "t" $field.ShapeRef.Shape.MemberRef "" "krTarget" 1 }}
		res = append(res, t)
	}
	return res
}

// resourceTargetsFromSDKTargets converts the given AWS service targets to Kubernetes resource targets
func resourceTargetsFromSDKTargets(
	targets []*svcsdk.Target,
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
