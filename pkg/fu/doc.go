/*
Package fu (File Utilities) implements utility functions
for working with file system items.
Features:
	* unified Copy function for copying file system items of any type;

References:

Pipes: https://www.youtube.com/watch?v=Mqb2dVRe0uo;

	+ Pipe is a sort of memory file for IPC;
	+ It can only be established between a process and its
	  subprocess(es);
	+ Any side can read or write to it and has to close its
	  read and write descriptors;

Named pipes: https://www.youtube.com/watch?v=2hba3etpoJg;

	+ Named pipes are also called `FIFO`s;
	+ Related ...nix command that creates a named pipe: `mkfifo`;
	+ Allows for communication between any processes with sufficient permissions;

*/
package fu
