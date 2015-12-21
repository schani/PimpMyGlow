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

### Fill

With time arithmetic we can run loops a specific number of iterations
that we arrive at by dividing the total duration we want by the
duration of one iteration.  That usually results in some amount of
time at the end of the loop to be unfilled because it's not enough for
a whole loop iteration, but it might be noticeable.

We can instead let the compiler calculate the number of iterations
that fit into the desired duration, and then insert one more iteration
up to the point that it fills in the rest.  For example:

	FILL,43
	    L,100000
		    COLOR,white
			D,5
			COLOR,black
			D,5
		E
	E

will shorten the number of iterations of the loop within the `FILL` to
fit into 43, namely to 4 iterations.  There is now a duration of 3
left to fill, so it will start adding commands from the loop until 3
is full, so it'll produce

    L,4
	    COLOR,white
		D,5
		COLOR,black
		D,5
	E
	COLOR,white
	D,3

`D` and `RAMP` commands will be shortened by `FILL` to produce the
correct fit.  This might be problematic in the case of `RAMP`.

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
subroutine will be inserted for the label, and filled to the length of
the label.  Usually the subroutine will consist of a loop, so you'll
probably want to choose an unrealistically big number of iterations so
that it gets shortened.  For example:

     DEFSUB,blink
	    L,1000000
		    COLOR,white
			D,5
			COLOR,black
			D,5
		E
	ENDSUB

Then if a label of duration `77` is named

    blink

it will produce the code

	L,7
		COLOR,white
		D,5
		COLOR,black
		D,5
	E
	COLOR,white
	D,5
	COLOR,black
	D,2

Within the subroutine, the variable `duration` stands for the duration
of the label.  This might be useful if you want to produce a specific
number of blinks, no matter how long or short the label is.

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

    C1,3,5:RAMP:black:white:black

will ramp only clubs 1, 3, and 5 from black to white to black
again.
