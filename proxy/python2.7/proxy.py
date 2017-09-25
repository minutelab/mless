# Copyright (C) 2017 MinLab Ltd.

from __future__ import print_function

import json
import os
import urllib2
import socket

def simplify(obj):
    keys = getattr(obj,'__slots__',None) or getattr(obj,'__dict__',None)
    if keys == None:
        return obj
    return {key: simplify(getattr(obj, key)) for key in keys}

def proxy(event, context, server=None, funcName=None):
    server = server or os.environ.get("MLESS_SERVER")
    if server == None:
        raise Exception("No server defined")

    funcName = funcName or os.environ.get("MLESS_FUNCNAME")

    env = dict(os.environ.items())
    if funcName != None:
        env["AWS_LAMBDA_FUNCTION_NAME"] = funcName

    req = {
       "event": event,
       "context": simplify(context),
       "remaining": context.get_remaining_time_in_millis(),
       "env": env
    }

    print("sending request to " + server)
    try:
        f = urllib2.urlopen(urllib2.Request(server+"/invoke", json.dumps(req), {'Content-Type': 'application/json'}))
    except urllib2.HTTPError as e:
        content=e.read()
        raise StandardError("ServerError: %s: %s" % (e.code, content))

    response = f.read()
    f.close()
    print("response:" +response)
    return response
