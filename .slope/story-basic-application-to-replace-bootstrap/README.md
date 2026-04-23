+++
_id = story-1
+++
# Basic application to replace bootstrap.

We have bootstrap code which implementes add and prompt,
but lacking _any_ testing, proper mergin and templating. 
While we can postpone templating, we highly likely can
start to use bootstrap for add but first to implement
proper prompt merging.

We have to identify the cmd/slope/main.go structure,
needed packages (it should be a cobra) and start to pile up
toward using its implementation of 'prompt'

We do not do it on once shot but in seriese of scope limited session
according to following