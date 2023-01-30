# Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You may
# not use this file except in compliance with the License. A copy of the
# License is located at
#
# 	 http://aws.amazon.com/apache2.0/
#
# or in the "license" file accompanying this file. This file is distributed
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.

"""Integration tests for the EventBridge bus API.
"""

import pytest
import time
import logging

from acktest.resources import random_suffix_name
from acktest.k8s import resource as k8s
from acktest.k8s import condition as condition
from acktest import tags
from e2e import service_marker, CRD_GROUP, CRD_VERSION, load_eventbridge_resource
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e.tests.helper import EventBridgeValidator

RESOURCE_PLURAL = "eventbuses"

CREATE_WAIT_AFTER_SECONDS = 10
UPDATE_WAIT_AFTER_SECONDS = 10
DELETE_WAIT_AFTER_SECONDS = 10

@pytest.fixture(scope="module")
def eventbridge_bus():
        resource_name = random_suffix_name("ack-test-bus", 24)

        replacements = REPLACEMENT_VALUES.copy()
        replacements["BUS_NAME"] = resource_name

        # Load EventBus CR
        resource_data = load_eventbridge_resource(
            "eventbus",
            additional_replacements=replacements,
        )
        logging.debug(resource_data)

        # Create k8s resource
        ref = k8s.CustomResourceReference(
            CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL,
            resource_name, namespace="default",
        )
        k8s.create_custom_resource(ref, resource_data)
        cr = k8s.wait_resource_consumed_by_controller(ref)

        assert cr is not None
        assert k8s.get_resource_exists(ref)

        time.sleep(CREATE_WAIT_AFTER_SECONDS)

        cr = k8s.wait_resource_consumed_by_controller(ref)

        yield (ref, cr)

        try:
            _, deleted = k8s.delete_custom_resource(ref, 3, 10)
            assert deleted
        except:
            pass


@service_marker
@pytest.mark.canary
class TestEventBus:
    def test_create_delete(self, eventbridge_client, eventbridge_bus):
        (ref, cr) = eventbridge_bus
        event_bus_name = cr["spec"]["name"]

        # Check eventbridge Bus exists
        eventbridge_validator = EventBridgeValidator(eventbridge_client)
        assert eventbridge_validator.event_bus_exists(event_bus_name)

        # Delete k8s resource
        _, deleted = k8s.delete_custom_resource(ref)
        assert deleted

        time.sleep(DELETE_WAIT_AFTER_SECONDS)

        # Check eventbridge Bus doesn't exist
        assert not eventbridge_validator.event_bus_exists(event_bus_name)

    def test_update(self, eventbridge_client, eventbridge_bus):
        (ref, cr) = eventbridge_bus
        event_bus_name = cr["spec"]["name"]

        # Check eventbridge Bus exists
        eventbridge_validator = EventBridgeValidator(eventbridge_client)
        assert eventbridge_validator.event_bus_exists(event_bus_name)

        event_bus_arn = cr["status"]["ackResourceMetadata"]["arn"]
        event_bus_tags = eventbridge_validator.get_resource_tags(event_bus_arn)
        tags.assert_ack_system_tags(
            tags=event_bus_tags,
        )
        tags_dict = tags.to_dict(
            cr["spec"]["tags"],
            key_member_name = 'key',
            value_member_name = 'value'
        )
        tags.assert_equal_without_ack_tags(
            actual=tags_dict,
            expected=event_bus_tags,
        )

        cr = k8s.wait_resource_consumed_by_controller(ref)

        # Update cr
        cr["spec"]["tags"] =  [
            {
                "key": "key",
                "value": "value-updated"
            }
        ]

        # Patch k8s resource
        k8s.patch_custom_resource(ref, cr)
        time.sleep(UPDATE_WAIT_AFTER_SECONDS)

        event_bus_tags = eventbridge_validator.get_resource_tags(event_bus_arn)
        tags.assert_ack_system_tags(
            tags=event_bus_tags,
        )
        tags_dict = tags.to_dict(
            cr["spec"]["tags"],
            key_member_name = 'key',
            value_member_name = 'value'
        )
        tags.assert_equal_without_ack_tags(
            actual=tags_dict,
            expected=event_bus_tags,
        )

        # Delete k8s resource
        _, deleted = k8s.delete_custom_resource(ref)
        assert deleted

        time.sleep(DELETE_WAIT_AFTER_SECONDS)

        # Check eventbridge Bus doesn't exist
        assert not eventbridge_validator.event_bus_exists(event_bus_name)