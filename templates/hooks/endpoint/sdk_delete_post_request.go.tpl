// always requeue if API call succeeded due to eventually consistent state
// transitions
if err == nil {
	return r, requeueWaitWhileDeleting
}
