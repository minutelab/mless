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

def proxy(event, context, funcName=None):
    handler = os.environ.get("MLESS_HANDLER")
    if handler == None:
        raise Exception("No handler defined")

    ctx = simplify(context)
    if funcName != None:
        ctx['function_name'] = funcName

    req = {
       "event": event,
       "context": ctx,
       "remaining": context.get_remaining_time_in_millis(),
       "env": dict(os.environ.items())
    }

    print("sending request to " + handler)
    print("request is "+ json.dumps(req))
    f = urllib2.urlopen(urllib2.Request(handler+"/invoke", json.dumps(req), {'Content-Type': 'application/json'}))
    response = f.read()
    f.close()
    print("response:" +response)
    return response
