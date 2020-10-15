# `profile-make` - A profiler for GNU Make

## Usage

Instead of running

   ```console
   $ make MAKE_ARGS
   ```

run

   ```console
   $ profile-make run --output-file=profile.json -- make MAKE_ARGS
   ```

Then, visualize what happened with

   ```console
   $ profile-make visualize <profile.json >profile.svg
   ```

## Limitations / gotchas

### Setting `SHELL`

If your Makefile sets `SHELL`, you'll need to adjust that a touch.
Instead of writing

   ```Makefile
   SHELL = myshell
   ```

write

   ```Makefile
   profile-make.SHELL = myshell
   SHELL = $(profile-make.SHELL)
   ```

### Pattern-rules with multiple outputs

It has trouble connecting nodes in the DAG for pattern rules with
multiple outputs.  For example:

   ```Makefile
   # source files:
   #  - program.c
   #  - grammar.y

   all: program
   .PHONY: all

   program: program.o grammar.tab.o
   	cc -o $@ $^

   program.o: grammar.tab.h

   %.o: %.c
   	cc -c -o $@ $<

   %.tab.h %.tab.c: %.y
   	bison -d $<
   ```

Make sees this DAG as:

                                     (all)
                                       |
                                   (program)
                                       |
                     +---------------------------------------+
                     | cc -o program program.o grammar.tab.o |
                     +---------------------------------------+
                                  |        |
                     ,------------'        `----------------,
                     |                                      |
                 (program.o)                          (grammar.tab.o)
                     |                                      |
       +------------------------------+  +--------------------------------------+
       | cc -c -o program.o program.c |  | cc -c -o grammar.tab.o grammar.tab.c |
       +------------------------------+  +--------------------------------------+
            |        |                                      |
            | (grammar.tab.h)                        (grammar.tab.c)
            |        |                                      |
            |        `------------,        ,----------------'
            |                     |        |
        (program.c)         +--------------------+
                            | bison -d grammar.y |
                            +--------------------+
                                       |
                                  (grammar.y)

However, `profile-make` will only see one of the legs coming off of
`bison -d grammar.y`; either

                                     (all)
                                       |
                                   (program)
                                       |
                     +---------------------------------------+
                     | cc -o program program.o grammar.tab.o |
                     +---------------------------------------+
                                  |        |
                     ,------------'        `----------------,
                     |                                      |
                 (program.o)                          (grammar.tab.o)
                     |                                      |
       +------------------------------+  +--------------------------------------+
       | cc -c -o program.o program.c |  | cc -c -o grammar.tab.o grammar.tab.c |
       +------------------------------+  +--------------------------------------+
            |        |                                      |
            | (grammar.tab.h)                        (grammar.tab.c)
            |                                               |
            |                              ,----------------'
            |                              |
        (program.c)         +--------------------+
                            | bison -d grammar.y |
                            +--------------------+
                                       |
                                  (grammar.y)

or

                                     (all)
                                       |
                                   (program)
                                       |
                     +---------------------------------------+
                     | cc -o program program.o grammar.tab.o |
                     +---------------------------------------+
                                  |        |
                     ,------------'        `----------------,
                     |                                      |
                 (program.o)                          (grammar.tab.o)
                     |                                      |
       +------------------------------+  +--------------------------------------+
       | cc -c -o program.o program.c |  | cc -c -o grammar.tab.o grammar.tab.c |
       +------------------------------+  +--------------------------------------+
            |        |                                      |
            | (grammar.tab.h)                        (grammar.tab.c)
            |        |
            |        `------------,
            |                     |
        (program.c)         +--------------------+
                            | bison -d grammar.y |
                            +--------------------+
                                       |
                                  (grammar.y)
