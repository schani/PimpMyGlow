# Pimp my Glow

This is an extension of the language Aerotech uses programming their glow clubs.
See [their manual](http://www.aerotechprojects.com/PDF/user-guide-V2.3.pdf) for
a description of the basic language.

This program is a simple compiler for the extended language, producing programs
that work with Aerotech's tools.

## Extensions

### Color names

Colors can be given names which can be used in the `C` and `RAMP`
commands, like so:

    COLOR,white,255,255,255
	COLOR,red,255,0,0,0
	
	C,red
	RAMP,white,100

Colors can also be modified with a percentage.  Given the color
definitions from above, this code

	C,white 20%

would produce

	C,50,50,50

### Multiple clubs

Instead of having to write a separate file for each club, we provide
a `CLUBS` command that lets you write commands that apply only to
specific clubs.  For example:

    C,black
	CLUBS,1,3,5
	    RAMP,red,100
	E
	CLUBS,2,4
		RAMP,white,100
	E

ramps to red for clubs 1, 3, and 5, and to white for clubs 2, 4.  Commands
that are given outside of `CLUBS` apply to all clubs.

### Absolute time

Instead of having to manually keep track of time we provide the command
`TIME` which produces a `D` that jumps ahead to an absolute time stamp.
For example, the program

	C,black
	D,100
	C,white
	TIME,1000

produces

	C,0,0,0
	D,100
	C,255,255,255
	D,900

### Labels from Audacity

You can mark sections in an audio file with the
[Audacity audio editor](http://audacityteam.org/) with "labels" and
then use the names of those labels in programs instead of putting in
numeric time stamps.

For example, let's say we have defined a `drums` label in Audacity and
we want the clubs to glow white as long as `drums` is active:

    C,black
	TIME,drums
	C,white
	TIME,-drums

A minus sign before the name of a label stands for the end time stamp
of the label, so the first `TIME` command in this example jumps to the
start of `drums`, and the second one to the end.

Another way to write this is to use the ampersand sign, which stands
for the duration of a label:

	C,black
	TIME,drums
	C,white
	D,&drums

### Time arithmetic

You can do simple arithmetic with time.  Right now only division is
supported.

Let's say we want the clubs to blink 10 times while `drums` is active:

    C,black
	TIME,drums
	L,10
	    C,white
		D,&drums/20
		C,black
		D,&drums,20
	E

In each iteration of the loop the clubs glow white for a 20th of the
total `drums` duration, then black for another 20th.

Note that because the resolution of the timer is only a hundredth
of a second, the total loop duration might be somewhat less than the
duration of `drums`, especially if you use a large number of iterations.

## Timeline

Another way to produce a program is to use labels in Audacity to mark
and specify colors or subroutines for all or specific clubs.  The
following kinds of labels can be used:

### Colors

The label name can be the name of a color, optionally with a percentage.
The color has to be defined in the `glo` file.  Example:

    red 50%

### Subroutines

If the label name is the name of a subroutine, the code for that
subroutine will be inserted for the label.  Within the subroutine,
the variable `duration` stands for the duration of the label.  For
example, if this is defined in the `glo` file:

     DEFSUB,blink
	    L,duration/10
		    COLOR,white
			D,5
			COLOR,black
			D,5
		E
	ENDSUB

then if a label of duration `70` is named

    blink

it will produce the code

	L,7
		COLOR,white
		D,5
		COLOR,black
		D,5
	E

Make sure that the duration of the subroutine is not longer than
the duration of the label, or you might get an error, or at least
unexpected results.

### Ramps

A label name of the form

    RAMP:C1:C2:...:Cn

will produce a series of ramps starting with `C1` to `C2`, through
to `Cn`.  The duration of the ramps will be evenly distributed to
fill the duration of the label.  At least two colors must be
specified.  Percentages can be used.  For example:

    RAMP:black:red 50%:white:black

will produce a ramp from black to half red, then to white, then
back to black.

### Specifying clubs

A label can be prefixed with something of the form

    C1,2,...,n:

to specify which clubs the label applies to.  If there is no
such specification, the label applies to all clubs.  For example:

    C1,3,5:RAMP:black,white,black

will ramp only clubs 1, 3, and 5 from black to white to black
again.
