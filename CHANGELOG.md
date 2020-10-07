# Changelog
All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

## [v2.0.2](https://github.com/imgix/imgix-go/compare/2.0.1...2.0.2) - October 6, 2020

### Changes
The changes in this release have been made to address a GoDocs issue. The only changes have been in this file and to bump the release version patch level.

## [v2.0.1](https://github.com/imgix/imgix-go/compare/2.0.0...2.0.1) - September 28, 2020

### Changes
The changes made in this release have been primarily cosmetic. Prior to this release the contents of `v2/` were duplicated in the project root. Now, that duplication has been eliminated and the Makefile has been updated accordingly (for testing).

- refactor: remove duplicate v2 files from project root ([#11](https://github.com/imgix/imgix-go/pull/11))
- build: update Makefile to test v2/ ([#11](https://github.com/imgix/imgix-go/pull/11))

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
go get github.com/imgix/imgix-go/v2
```

## [v2.0.0](https://github.com/imgix/imgix-go/compare/1.0.0...2.0.0) - September 24, 2020

### Breaking Changes
imgix-go has undergone a complete rewrite in order to reach parity with the rest of [imgix's SDK](https://docs.imgix.com/libraries#client-libraries). Our team is excited to share these changes via a new major release -- v2.0.0.

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
go get github.com/imgix/imgix-go/v2
```
