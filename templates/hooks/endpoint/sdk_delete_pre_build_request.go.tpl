if endpointInMutatingState(r) {
	return r, requeueWaitUntilCanModify(r)
}
