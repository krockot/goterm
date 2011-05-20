#include <termios.h>
#include <sys/ioctl.h>
#include <pty.h>
#include <utmp.h>
#include <unistd.h>

typedef struct {
    int master;
    int slave;
    int result;
} openpty_result;

openpty_result goterm_openpty(struct termios* ios, struct winsize* size);
int goterm_get_window_size(int fd, struct winsize* sz);
int goterm_set_window_size(int fd, struct winsize* sz);
int goterm_fcntl(int, int, int);

