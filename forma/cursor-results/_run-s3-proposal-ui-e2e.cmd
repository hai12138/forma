@echo off
setlocal
set FORMA_LIVE_E2E=1
set FORMA_LIVE_BASE_URL=http://127.0.0.1:8888
set FORMA_UI_BASE_URL=http://127.0.0.1:3001
set FORMA_S3_E2E_RESUME=1
set MAX_REAL_MODEL_CALLS=0
cd /d d:\product\Forma\forma-workspace\coze-studio
node --test scripts/forma/s3-proposal-browser-ui-e2e.mjs
