# Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You may
# not use this file except in compliance with the License. A copy of the
# License is located at
#
#	 http://aws.amazon.com/apache2.0/
#
# or in the "license" file accompanying this file. This file is distributed
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.

"""Helper functions for EventBridge e2e tests
"""

import logging

class EventBridgeValidator:
    def __init__(self, eventbridge_client):
        self.eventbridge_client = eventbridge_client

    def get_event_bus(self, event_bus_name: str) -> dict:
        try:
            resp = self.eventbridge_client.describe_event_bus(
                Name=event_bus_name
            )
            return resp

        except Exception as e:
            logging.debug(e)
            return None

    def event_bus_exists(self, event_bus_name) -> bool:
        return self.get_event_bus(event_bus_name) is not None

    def get_resource_tags(self, resource_arn: str):
        resource_tags = self.eventbridge_client.list_tags_for_resource(
            ResourceARN=resource_arn,
        )
        return resource_tags['Tags']