# go-denvr

A Go SDK for interacting with Denvr Cloud SDK

[![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/denvrdata/go-denvr/CI.yml)](https://github.com/denvrdata/go-denvr/actions/workflows/CI.yml)
[![Coveralls](https://img.shields.io/coverallsCoverage/github/denvrdata/go-denvr)](https://coveralls.io/github/denvrdata/go-denvr?branch=main)
[![Swagger Validator](https://img.shields.io/swagger/valid/3.0?specUrl=https%3A%2F%2Fapi.cloud.denvrdata.com%2Fswagger%2Fv1%2Fswagger.json)](https://api.cloud.denvrdata.com/swagger/index.html)
[![go-denvr Docs](https://img.shields.io/badge/denvr-docs-%234493c5?style=flat)](https://pkg.go.dev/github.com/denvrdata/go-denvr)
[![Denvr Dataworks Docs](https://img.shields.io/badge/denvr_cloud-docs-%234493c5?style=flat)](https://docs.denvrdata.com/docs)
[![Denvr Dataworks Registration](https://img.shields.io/badge/denvr_cloud-registration-%234493c5?style=flat)](https://console.cloud.denvrdata.com/account/register-tenant)

## Why go-denvr?

It's a bit of a hack, so that the `go-` prefix can be excluded when using the package.
The following reddit [thread](https://www.reddit.com/r/golang/comments/r3as15/how_should_i_name_my_package_repository_when/) explains.

## Why a golang SDK?

We prioritised making `denvrpy` as many of our users are familiar with writing python code.
A golang SDK allows us to write a denvr terraform provider.
