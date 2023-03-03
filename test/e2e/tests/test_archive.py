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

"""Integration tests for the EventBridge Archive API.
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

RESOURCE_PLURAL = "archives"

CREATE_WAIT_AFTER_SECONDS = 10
UPDATE_WAIT_AFTER_SECONDS = 10
DELETE_WAIT_AFTER_SECONDS = 10

@pytest.fixture(scope="module")
def event_bus():
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
            CRD_GROUP, CRD_VERSION, "eventbuses",
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

@pytest.fixture(scope="module")
def archive(event_bus):
        resource_name = random_suffix_name("ack-test-archive", 24)
        _, eb_cr = event_bus

        replacements = REPLACEMENT_VALUES.copy()
        replacements["ARCHIVE_NAME"] = resource_name
        replacements["EVENT_SOURCE_ARN"] = eb_cr["status"]["ackResourceMetadata"]["arn"]

        # Load EventBus CR
        resource_data = load_eventbridge_resource(
            "archive",
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
class TestArchive:
    def test_crud(self, eventbridge_client, archive):
        (ref, cr) = archive
        archive_name = cr["spec"]["archiveName"]

        # Check archive exists
        eventbridge_validator = EventBridgeValidator(eventbridge_client)
        assert eventbridge_validator.archive_exists(archive_name)

        new_description = "new archive description"
        cr["spec"]["description"] =  new_description

        # Patch k8s resource
        k8s.patch_custom_resource(ref, cr)
        time.sleep(UPDATE_WAIT_AFTER_SECONDS)

        # Check description new value
        archive = eventbridge_validator.get_archive(archive_name)
        assert archive["Description"] == new_description

        # Delete k8s resource
        _, deleted = k8s.delete_custom_resource(ref)
        assert deleted

        time.sleep(DELETE_WAIT_AFTER_SECONDS)

        # Check archive doesn't exist
        assert not eventbridge_validator.archive_exists(archive_name)