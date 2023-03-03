if err = validateEndpointSpec(delta, desired.ko.Spec); err != nil {
	return nil, ackerr.NewTerminalError(err)
}

if endpointInMutatingState(latest) {
	return latest, requeueWaitWhileUpdating
}
