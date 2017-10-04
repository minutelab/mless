# Copyright (C) 2017 MinLab Ltd.

import json
import os
import urllib.error
import urllib.request
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
        f = urllib.request.urlopen(urllib.request.Request(server+"/invoke", json.dumps(req).encode('utf-8'), {'Content-Type': 'application/json'}))
    except urllib.error.HTTPError as e:
        content=e.read()
        raise RuntimeError("ServerError: %s: %s" % (e.code, content))

    response = f.read().decode("utf-8")
    f.close()
    print("response:" + response)
    try:
        res = json.loads(response)
        return res
    except Exception as e:
        print(e)
        return response
