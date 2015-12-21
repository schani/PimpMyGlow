# Check consistency

# UI: Save paths when quitting

# Document timeline

# Set default color

Right now it's black.

# Fill loops

Right now we can define subs and make them run
a specific number of iterations that we arrive at
by dividing the duration of the label by the duration
of one iteration.  That usually results in some amount
of time at the end of the label to be unfilled
because it's not enough for a whole loop iteration,
but it might be noticable.

We could instead let the compiler calculate the
number of iterations that fit into the label, and then
unroll one more iteration up to the point that it fills
in the rest.  For example, given

    DEFSUB,blink
	    L,100000
		    COLOR,white
			D,5
			COLOR,black
			D,5
		E
	ENDSUB

filling a label of duration 43 would produce

    L,4
	    COLOR,white
		D,5
		COLOR,black
		D,5
	E
	COLOR,white
	D,3
