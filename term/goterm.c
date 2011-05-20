#include <sys/ioctl.h>
#include <termios.h>
#include <stdlib.h>
#include "goterm.h"

openpty_result goterm_openpty(struct termios *ios, struct winsize *size)
{
	openpty_result result;
	result.result = openpty(&result.master, &result.slave, NULL, ios, size);
	return result;
}

int goterm_get_window_size(int fd, struct winsize *sz)
{
	return ioctl(fd, TIOCGWINSZ, sz);
}

int goterm_set_window_size(int fd, struct winsize *sz)
{
	return ioctl(fd, TIOCSWINSZ, sz);
}

int goterm_fcntl(int fd, int cmd, int arg)
{
	return fcntl(fd, cmd, arg);
}

