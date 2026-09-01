#!/bin/bash
export HURL_VARIABLE_ID=1
export HURL_VARIABLE_NAME="Foo"
export HURL_VARIABLE_BASE64_IMAGE=$(base64 -i ./images/qrcode.png)
hurl tests/add-post.hurl
