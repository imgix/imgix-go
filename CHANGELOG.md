# Changelog
All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

## [v2.0.0](https://github.com/imgix/imgix-go/compare/1.0.0...2.0.0) - September 23, 2020

### Breaking Changes
imgix-go has undergone complete rewrite in order to reach parity with the rest of [imgix's SDK](https://docs.imgix.com/libraries#client-libraries). Our team is excited to share these changes via a new major release -- v2.0.0.

### Features
* imgix [URL auto generation](https://github.com/imgix/imgix-go#usage)
  * HTTPS and HTTP support
  * [Token-secured URLs](https://docs.imgix.com/setup/securing-images#enabling-secure-urls)
* automatic [srcset generation](https://github.com/imgix/imgix-go#srcset-generation)
* customizable [fixed-width](https://github.com/imgix/imgix-go#fixed-width-images) srcset via [variable qualities](https://github.com/imgix/imgix-go#variable-quality)
* customizable [fluid-width](https://github.com/imgix/imgix-go#fluid-width-images) srcsets via the following:
  * [custom widths](https://github.com/imgix/imgix-go#custom-widths)
  * [width ranges](https://github.com/imgix/imgix-go#width-ranges)
  * [width tolerance](https://github.com/imgix/imgix-go#width-tolerance)
  * [target widths](https://github.com/imgix/imgix-go#width-tolerance)

### Install

```
go get github.com/imgix/imgix-go
```
