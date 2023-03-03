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

"""Integration tests for the EventBridge Endpoint API.
"""

import pytest
import time
import logging

from acktest.resources import random_suffix_name
from acktest.k8s import resource as k8s
from acktest.k8s import condition as condition
from acktest.aws.identity import get_region
from e2e import service_marker, CRD_GROUP, CRD_VERSION, load_eventbridge_resource
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e.tests.helper import EventBridgeValidator
from e2e.bootstrap_resources import get_bootstrap_resources

RESOURCE_PLURAL = "endpoints"

CREATE_WAIT_AFTER_SECONDS = 10
UPDATE_WAIT_AFTER_SECONDS = 10
DELETE_WAIT_AFTER_SECONDS = 10

def create_event_bus(aws_name: str, region: str):
    resource_name = random_suffix_name("ack-test-bus", 24)
    replacements = REPLACEMENT_VALUES.copy()
    replacements["BUS_NAME"] = aws_name

    # Load EventBus CR
    resource_data = load_eventbridge_resource(
        "eventbus",
        additional_replacements=replacements,
    )
    logging.debug(resource_data)

    # Create k8s resource
    resource_data["metadata"] = {
        "name": resource_name,
        "annotations": {"services.k8s.aws/region": region},
    }

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

    return ref, cr

@pytest.fixture(scope="module")
def bus_name():
    return random_suffix_name("bus", 24)

@pytest.fixture(scope="module")
def bus_uswest1(bus_name):
        ref, cr = create_event_bus(bus_name, "us-west-1")
        yield ref, cr

        try:
            _, deleted = k8s.delete_custom_resource(ref, 3, 10)
            assert deleted
        except:
            pass

@pytest.fixture(scope="module")
def bus_uswest2(bus_name):
        ref, cr = create_event_bus(bus_name, "us-west-2")
        yield ref, cr

        try:
            _, deleted = k8s.delete_custom_resource(ref, 3, 10)
            assert deleted
        except:
            pass

@pytest.fixture(scope="module")
def endpoint(bus_uswest1, bus_uswest2):
    _, eb_a = bus_uswest1
    _, eb_b = bus_uswest2

    resource_name = random_suffix_name("ack-test-endpoint", 24)

    resources = get_bootstrap_resources()
    replacements = REPLACEMENT_VALUES.copy()
    replacements["ENDPOINT_NAME"] = resource_name
    replacements["EVENT_BUS_ARN_A"] = eb_a["status"]["ackResourceMetadata"]["arn"]
    replacements["EVENT_BUS_ARN_B"] = eb_b["status"]["ackResourceMetadata"]["arn"]
    replacements["HEALTH_CHECK_LOCATION"] = f"arn:aws:route53:::healthcheck/{resources.EndpointHealthCheck.id}"

    # Load EventBus CR
    resource_data = load_eventbridge_resource(
        "endpoint",
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
class TestEndpoint:
    def test_crud(self, eventbridge_client, endpoint):
        (ref, cr) = endpoint
        endpoint_name = cr["spec"]["name"]

        # Check endpoint exists
        eventbridge_validator = EventBridgeValidator(eventbridge_client)
        assert eventbridge_validator.endpoint_exists(endpoint_name)

        new_description = "new endpoint description"
        cr["spec"]["description"] =  new_description

        # Patch k8s resource
        k8s.patch_custom_resource(ref, cr)
        time.sleep(UPDATE_WAIT_AFTER_SECONDS)

        # Check description new value
        endpoint = eventbridge_validator.get_endpoint(endpoint_name)
        assert endpoint["Description"] == new_description

        # Delete k8s resource
        _, deleted = k8s.delete_custom_resource(ref, 3, 10)
        assert deleted
        time.sleep(DELETE_WAIT_AFTER_SECONDS)

        # Check endpoint doesn't exist
        assert not eventbridge_validator.endpoint_exists(endpoint_name)