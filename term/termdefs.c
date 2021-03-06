#include <termios.h>
#include <sys/ioctl.h>

enum {
	$VEOF = VEOF,
	$VEOL = VEOL,
	$VERASE = VERASE,
	$VINTR = VINTR,
	$VKILL = VKILL,
	$VMIN = VMIN,
	$VQUIT = VQUIT,
	$VSTART = VSTART,
	$VSTOP = VSTOP,
	$VSUSP = VSUSP,
	$VTIME = VTIME
};

enum {
	$BRKINT = BRKINT,
	$ICRNL = ICRNL,
	$IGNBRK = IGNBRK,
	$IGNCR = IGNCR,
	$IGNPAR = IGNPAR,
	$INLCR = INLCR,
	$INPCK = INPCK,
	$ISTRIP = ISTRIP,
	$IUCLC = IUCLC,
	$IXANY = IXANY,
	$IXOFF = IXOFF,
	$IXON = IXON,
	$PARMRK = PARMRK
};

enum {
	$OPOST = OPOST,
	$OLCUC = OLCUC,
	$ONLCR = ONLCR,
	$OCRNL = OCRNL,
	$ONOCR = ONOCR,
	$ONLRET = ONLRET,
	$OFILL = OFILL,
	$NLDLY = NLDLY,
	$NL0 = NL0,
	$NL1 = NL1,
	$CRDLY = CRDLY,
	$CR0 = CR0,
	$CR1 = CR1,
	$CR2 = CR2,
	$CR3 = CR3,
	$TABDLY = TABDLY,
	$TAB0 = TAB0,
	$TAB1 = TAB1,
	$TAB2 = TAB2,
	$TAB3 = TAB3,
	$BSDLY = BSDLY,
	$BS0 = BS0,
	$BS1 = BS1,
	$VTDLY = VTDLY,
	$VT0 = VT0,
	$VT1 = VT1,
	$FFDLY = FFDLY,
	$FF0 = FF0,
	$FF1 = FF1
};

enum {
	$B0 = B0,
	$B50 = B50,
	$B75 = B75,
	$B110 = B110,
	$B134 = B134,
	$B150 = B150,
	$B200 = B200,
	$B300 = B300,
	$B600 = B600,
	$B1200 = B1200,
	$B1800 = B1800,
	$B2400 = B2400,
	$B4800 = B4800,
	$B9600 = B9600,
	$B19200 = B19200,
	$B38400 = B38400
};

enum {
	$CSIZE = CSIZE,
	$CS5 = CS5,
	$CS6 = CS6,
	$CS7 = CS7,
	$CS8 = CS8,
	$CSTOPB = CSTOPB,
	$CREAD = CREAD,
	$PARENB = PARENB,
	$PARODD = PARODD,
	$HUPCL = HUPCL,
	$CLOCAL = CLOCAL
};

enum {
	$ECHO = ECHO,
	$ECHOE = ECHOE,
	$ECHOK = ECHOK,
	$ECHONL = ECHONL,
	$ICANON = ICANON,
	$IEXTEN = IEXTEN,
	$ISIG = ISIG,
	$NOFLSH = NOFLSH,
	$TOSTOP = TOSTOP,
	$XCASE = XCASE
};

enum {
	$TCSANOW = TCSANOW,
	$TCSADRAIN = TCSADRAIN,
	$TCSAFLUSH = TCSAFLUSH
};

enum {
	$TCIFLUSH = TCIFLUSH,
	$TCIOFLUSH = TCIOFLUSH,
	$TCOFLUSH = TCOFLUSH,
	$TCIOFF = TCIOFF,
	$TCION = TCION,
	$TCOOFF = TCOOFF,
	$TCOON = TCOON
};

enum {
	$NCCS = NCCS
};

enum {
	$TIOCSCTTY = TIOCSCTTY
};

