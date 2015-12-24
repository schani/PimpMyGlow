# UI: Save paths when quitting

# Set default color

Right now it's black.

# Fill ramp properly

We'll have to extend the `RAMP` command to take the actualy duration and
also the "original" duration.  Then in a late pass we need to keep track
of the color of the club so we can calculate the color the `RAMP` would
have to be if it stopped short.
