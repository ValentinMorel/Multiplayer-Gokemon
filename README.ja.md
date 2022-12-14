![logo](./logo.png)

# ð worldwide
![Go](https://github.com/pokemium/worldwide/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/pokemium/worldwide)](https://goreportcard.com/report/github.com/pokemium/worldwide)
[![GitHub stars](https://img.shields.io/github/stars/pokemium/worldwide)](https://github.com/pokemium/worldwide/stargazers)
[![GitHub license](https://img.shields.io/github/license/pokemium/worldwide)](https://github.com/pokemium/worldwide/blob/master/LICENSE)

Goè¨èªã§æ¸ãããã²ã¼ã ãã¼ã¤ã«ã©ã¼ã¨ãã¥ã¬ã¼ã¿ã§ãã  

å¤ãã®ROMãåé¡ãªãåä½ãããµã¦ã³ãæ©è½ãã»ã¼ãæ©è½ãªã©å¹åºãæ©è½ãåããã¨ãã¥ã¬ã¼ã¿ã§ãã

<img src="https://imgur.com/ZlrXAW9.png" width="320px"> <img src="https://imgur.com/xVqjkrk.png" width="320px"><br/>
<img src="https://imgur.com/E7oob9c.png" width="320px"> <img src="https://imgur.com/nYpkH95.png" width="320px">

## ð© ãã®ã¨ãã¥ã¬ã¼ã¿ã®ç¹å¾´ & ä»å¾å®è£äºå®ã®æ©è½
- [x] 60fpsã§åä½
- [x] [cpu_instrs](https://github.com/retrio/gb-test-roms/tree/master/cpu_instrs) ã¨ [instr_timing](https://github.com/retrio/gb-test-roms/tree/master/instr_timing)ã¨ãããã¹ãROMãã¯ãªã¢ãã¦ãã¾ã
- [x] å°ãªãCPUä½¿ç¨ç
- [x] ãµã¦ã³ãã®å®è£
- [x] ã²ã¼ã ãã¼ã¤ã«ã©ã¼ã®ã½ããã«å¯¾å¿
- [x] WindowsãLinuxãªã©æ§ããªãã©ãããã©ã¼ã ã«å¯¾å¿
- [x] MBC1, MBC2, MBC3, MBC5ã«å¯¾å¿
- [x] RTCã®å®è£
- [x] ã»ã¼ãæ©è½ããµãã¼ã(å¾ãããsavãã¡ã¤ã«ã¯å®æ©ãBGBãªã©ã®ä¸è¬çãªã¨ãã¥ã¬ã¼ã¿ã§å©ç¨ã§ãã¾ã)
- [x] ã¦ã£ã³ãã¦ã®ç¸®å°æ¡å¤§ãå¯è½
- [x] HTTPãµã¼ãã¼API
- [ ] ãã©ã°ã¤ã³æ©è½
- [ ] [Libretro](https://docs.libretro.com/) APIã®ãµãã¼ã
- [ ] ã­ã¼ã«ã«ãããã¯ã¼ã¯åã§ã®éä¿¡ãã¬ã¤
- [ ] ã°ã­ã¼ãã«ãããã¯ã¼ã¯åã§ã®éä¿¡ãã¬ã¤
- [ ] SGBã®ãµãã¼ã
- [ ] ã·ã§ã¼ãã®ãµãã¼ã

## ð® ä½¿ãæ¹

[ãã](https://github.com/pokemium/worldwide/releases)ããå®è¡ãã¡ã¤ã«ããã¦ã³ã­ã¼ãããå¾ãæ¬¡ã®ããã«èµ·åãã¾ãã

```sh
./worldwide "***.gb" # ãããã¯ `***.gbc`
```

## ð HTTPãµã¼ãã¼

`worldwide`ã¯HTTPãµã¼ãã¼ãååãã¦ãããã¦ã¼ã¶ã¼ã¯HTTPãªã¯ã¨ã¹ããéãã¦ `worldwide`ã«ãã¾ãã¾ãªæç¤ºãåºããã¨ãå¯è½ã§ãã

[ãµã¼ãã¼ãã­ã¥ã¡ã³ã](./server/README.md)ãåç§ãã¦ãã ããã

## ð¨ ãã«ã

ã½ã¼ã¹ã³ã¼ããããã«ããããæ¹åãã§ãã

requirements
- Go 1.16
- make

```sh
make build                              # Windowsãªã `make build-windows`
./build/darwin-amd64/worldwide "***.gb" # Windowsãªã `./build/windows-amd64/worldwide.exe "***.gb"`
```

## ð ã³ãã³ã

| ã­ã¼å¥å             | ã³ãã³ã      |
| -------------------- | ------------- |
| <kbd>&larr;</kbd>    | &larr; ãã¿ã³ |
| <kbd>&uarr;</kbd>    | &uarr; ãã¿ã³ |
| <kbd>&darr;</kbd>    | &darr; ãã¿ã³ |
| <kbd>&rarr;</kbd>    | &rarr; ãã¿ã³ |
| <kbd>X</kbd>         | A ãã¿ã³      |
| <kbd>Z</kbd>         | B ãã¿ã³      |
| <kbd>Enter</kbd>     | Start ãã¿ã³  |
| <kbd>Backspace</kbd> | Select ãã¿ã³ |
