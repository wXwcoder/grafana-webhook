#!/usr/bin/env bash
nohub ./webhook 8081 "feishu" https://open.feishu.cn/open-apis/bot/v2/hook/xxx >./webhook.log 2>&1 &
