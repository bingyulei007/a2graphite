#!/usr/bin/env python2
# -*- coding: utf-8 -*-

"""Collect sample messages and store as JSON for human reading.
"""

import os
import socket
import msgpack
import simplejson as json

sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
sock.bind(("0.0.0.0", 4952))

base_dir = os.path.dirname(os.path.abspath(__file__))
sample_dir = os.path.join(base_dir, "samples")

def redact_by_key(msg, key, value):
    segments = key.split(".")
    for k in segments[:-1]:
        try:
            msg = msg[k]
        except:
            try:
                msg = msg[int(k)]
            except:
                return
    k = segments[-1]
    if k in msg:
        msg[k] = value

def redact_msg(msg):
    redact_by_key(msg, "resource_metadata.display_name", "Redacted Instance Name ******")
    redact_by_key(msg, "resource_metadata.image.name", "Redacted Image Name ******")
    redact_by_key(msg, "resource_metadata.image.links.0.href", "Redacted url ******")
    redact_by_key(msg, "resource_metadata.flavor.links.0.href", "Redacted url ******")
    redact_by_key(msg, "resource_metadata.image_ref_url", "Redacted url ******")

while True:
    data, addr = sock.recvfrom(2048)
    msg = msgpack.unpackb(data)
    counter_name = msg.get("counter_name")
    sample_file = os.path.join(sample_dir, counter_name + ".json")
    if not os.path.exists(sample_file):
        redact_msg(msg)
        with open(sample_file, "w+") as fp:
            json.dump(msg, fp, indent=4)
        print "Collected sample: %s" % counter_name
