#!/bin/bash

act push \
  --bind \
  -s BOOK_GITHUB_TOKEN=ghp_fakeToken123456 \
  -e ACTIONS_RUNTIME_TOKEN=ghp_fakeToken123456 \
  --eventpath .github/test-event.json