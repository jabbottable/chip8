# chip8

chip8 is a emulator for the [CHIP-8](https://en.wikipedia.org/wiki/CHIP-8) programming language. The project is still in development.

## Installation

```bash
go build
```

## Usage

In a 64 character width terminal run:

```bash
./chip8 -f [file]
```

## TODO

- [ ] Implement full list of CHIP-8 opcodes
- [ ] Allow dymanic screen sizes
- [x] Add command line flag for ROM file
- [ ] Add proper tests
- [ ] Improve usage of Go language features
- [ ] Don't read sprites from file
