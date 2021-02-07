#!/usr/bin/env python
'''
API health check service
'''

import os
import sys
import json
import logging
from pprint import pformat

# Hack to use dependencies from lib directory
BASE_PATH = os.path.dirname(__file__)
sys.path.append(BASE_PATH + "/lib")

LOGGER = logging.getLogger(__name__)
logging.getLogger().setLevel(logging.INFO)

def response(status=200, headers=None, body=''):
    '''
    '''
    if not body:
        return {'statusCode': status}

    if headers is None:
        headers = {'Content-Type': 'application/json'}

    return {
        'statusCode': status,
        'headers': headers,
        'body': json.dumps(body)
    }

def lambda_handler(event, context):
    '''
    This will only confirm we are alive
    '''
    LOGGER.info("%s", pformat({"Context" : vars(context), "Request": event}))
    return response(status=200)


if __name__ == '__main__':
    # Do nothing if executed as a script
    pass

