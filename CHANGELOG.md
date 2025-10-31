# Changelog

## [0.5.1](https://github.com/mikesmitty/rp24-dcc-decoder/compare/v0.5.0...v0.5.1) (2025-10-31)


### Bug Fixes

* **deps:** update golang.org/x/exp digest to 42675ad ([#55](https://github.com/mikesmitty/rp24-dcc-decoder/issues/55)) ([8c79fb0](https://github.com/mikesmitty/rp24-dcc-decoder/commit/8c79fb0a1e2817a186f70df58597ea7953df09a9))
* **deps:** update golang.org/x/exp digest to 51f8813 ([#51](https://github.com/mikesmitty/rp24-dcc-decoder/issues/51)) ([7244902](https://github.com/mikesmitty/rp24-dcc-decoder/commit/724490245d16994d8783d3452a990d92a2ad6594))
* **deps:** update golang.org/x/exp digest to 8b4c13b ([#56](https://github.com/mikesmitty/rp24-dcc-decoder/issues/56)) ([41abde0](https://github.com/mikesmitty/rp24-dcc-decoder/commit/41abde0aa6662b7aac0c94eba1e1472ebc4626d6))
* **deps:** update golang.org/x/exp digest to a4bb9ff ([#64](https://github.com/mikesmitty/rp24-dcc-decoder/issues/64)) ([0dec770](https://github.com/mikesmitty/rp24-dcc-decoder/commit/0dec7704e9ae3c3ff21456ee4219fc70a096a228))
* **deps:** update golang.org/x/exp digest to d2f985d ([#63](https://github.com/mikesmitty/rp24-dcc-decoder/issues/63)) ([2d0001a](https://github.com/mikesmitty/rp24-dcc-decoder/commit/2d0001a6984248e023278804a349cace2932afc7))
* **deps:** update golang.org/x/exp digest to df92998 ([#61](https://github.com/mikesmitty/rp24-dcc-decoder/issues/61)) ([d0d50a4](https://github.com/mikesmitty/rp24-dcc-decoder/commit/d0d50a4dca40c667b3305ba84ea4512c51026bc5))

## [0.5.0](https://github.com/mikesmitty/rp24-dcc-decoder/compare/v0.4.0...v0.5.0) (2025-08-02)


### Features

* add early stage backemf motor control ([a25211d](https://github.com/mikesmitty/rp24-dcc-decoder/commit/a25211db5da32ed0f17aac7e8010aa097b95d403))
* add watchdog for signal cutout ([a25211d](https://github.com/mikesmitty/rp24-dcc-decoder/commit/a25211db5da32ed0f17aac7e8010aa097b95d403))


### Bug Fixes

* add watchdog timer ([11b6576](https://github.com/mikesmitty/rp24-dcc-decoder/commit/11b657602e25fd6cb13ee2e36697979c7ffdf19e))
* defend against panics from invalid messages ([ffdcd00](https://github.com/mikesmitty/rp24-dcc-decoder/commit/ffdcd002d98769922d42dc0cb6c4ae6520a1eed1))
* **deps:** update golang.org/x/exp digest to 645b1fa ([#48](https://github.com/mikesmitty/rp24-dcc-decoder/issues/48)) ([313a73c](https://github.com/mikesmitty/rp24-dcc-decoder/commit/313a73c76a07aca35d5ec08aace8f12d40190b84))

## [0.4.0](https://github.com/mikesmitty/rp24-dcc-decoder/compare/v0.3.0...v0.4.0) (2025-07-03)


### Features

* detect right rail instead of VDC ([ddb6ea3](https://github.com/mikesmitty/rp24-dcc-decoder/commit/ddb6ea383b4f5dc82e747a8a50b871392123b382))


### Bug Fixes

* fix cv write bugs ([bde219d](https://github.com/mikesmitty/rp24-dcc-decoder/commit/bde219d6429ba6914b5c6cf0a966d8f1f6285c03))

## [0.3.0](https://github.com/mikesmitty/rp24-dcc-decoder/compare/v0.2.0...v0.3.0) (2025-07-02)


### Features

* add basic ack functionality ([f5d5ae6](https://github.com/mikesmitty/rp24-dcc-decoder/commit/f5d5ae6a7456e60b06a05bcb1aa5879a4456d067))
* overhaul motor controller ([99d4ee8](https://github.com/mikesmitty/rp24-dcc-decoder/commit/99d4ee804c4e48eec4de7833bd355d6945b4c168))


### Bug Fixes

* add missing import ([c64d183](https://github.com/mikesmitty/rp24-dcc-decoder/commit/c64d183ea62e0248e458cf550c6ac9474b331468))
* enable motor control lockout mutex ([a0d5d45](https://github.com/mikesmitty/rp24-dcc-decoder/commit/a0d5d454318a02a385601f88090ae5ff6baa3785))

## [0.2.0](https://github.com/mikesmitty/rp24-dcc-decoder/compare/v0.1.0...v0.2.0) (2025-03-23)


### Features

* add motor control ([26f15c5](https://github.com/mikesmitty/rp24-dcc-decoder/commit/26f15c5dfaf2c012a957807f6a8c6f23504239b2))
* update go to 1.24.1 ([fa4d50e](https://github.com/mikesmitty/rp24-dcc-decoder/commit/fa4d50ecfa9deb2a11a8e6c42269f53fc977b5fc))


### Bug Fixes

* correct function offset indexes ([ea1fa88](https://github.com/mikesmitty/rp24-dcc-decoder/commit/ea1fa88235f76972585350e53caf278e261eab66))
* disable directional outputs when switching direction ([26f15c5](https://github.com/mikesmitty/rp24-dcc-decoder/commit/26f15c5dfaf2c012a957807f6a8c6f23504239b2))
* embed tagged version at build time ([009c713](https://github.com/mikesmitty/rp24-dcc-decoder/commit/009c713c4f9f1269592b5a993daec2d3cb00894c))

## [0.1.0](https://github.com/mikesmitty/rp24-dcc-decoder/compare/v0.0.1...v0.1.0) (2025-03-23)


### Features

* add function/output mapping and general cleanup ([d0cb788](https://github.com/mikesmitty/rp24-dcc-decoder/commit/d0cb788c2bafbf0a5f7e7699384653fa242ed265))
* add initial rough draft packages ([94049e1](https://github.com/mikesmitty/rp24-dcc-decoder/commit/94049e1630e36b6ee2894b043c71bf0edd1b20ed))


### Bug Fixes

* clean up motor CV definitions ([ddde344](https://github.com/mikesmitty/rp24-dcc-decoder/commit/ddde344ea2ec8ca0b557756d1d70a4cd0f5d16f2))
* cleanup and recategorize future changes ([19ba12f](https://github.com/mikesmitty/rp24-dcc-decoder/commit/19ba12f27c25d43e626b6c9f27d5b2940e60cd3c))
* **deps:** update module github.com/tinygo-org/pio to v0.2.0 ([#29](https://github.com/mikesmitty/rp24-dcc-decoder/issues/29)) ([1cd45db](https://github.com/mikesmitty/rp24-dcc-decoder/commit/1cd45db01dc1dcfaa50f6969206e61af102fbd75))
* handle double uno-reverse direction of travel swaps ([be3c416](https://github.com/mikesmitty/rp24-dcc-decoder/commit/be3c4161fb6095494123a71e18ab77e418a75bd7))
* implement CV21/CV22 consist function masks ([5bd3659](https://github.com/mikesmitty/rp24-dcc-decoder/commit/5bd3659a5e1a80b0e5bbf3bf347243e021f56986))
* load new index when CV31/CV32 are set ([5bd3659](https://github.com/mikesmitty/rp24-dcc-decoder/commit/5bd3659a5e1a80b0e5bbf3bf347243e021f56986))
* make dcc functions map to outputs ([8c448f2](https://github.com/mikesmitty/rp24-dcc-decoder/commit/8c448f22cd591466faf243c840eda684ffdd44ae))
* misc cleanup and add CVs ([d987945](https://github.com/mikesmitty/rp24-dcc-decoder/commit/d98794585a2216fc0b2ad298861d51faa18c7f80))
* squash various bugs ([8c448f2](https://github.com/mikesmitty/rp24-dcc-decoder/commit/8c448f22cd591466faf243c840eda684ffdd44ae))
* undo copy/paste bugs ([cc430f6](https://github.com/mikesmitty/rp24-dcc-decoder/commit/cc430f6f70852d3c441f914362f6252d59be9529))

## 0.0.1 (2025-02-16)
