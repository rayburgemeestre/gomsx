#include "SDL.h"

#include <X11/Xlib.h>

int window_id = 0;
int width = 0;
int height = 0;

int initSDL() {
	if (SDL_Init(SDL_INIT_AUDIO | SDL_INIT_VIDEO) != 0) {
		SDL_Log("Unable to initialize SDL: %s", SDL_GetError());
		return 1;
	}
	return 0;
}

int initWidthAndHeight()
{
    char *env = getenv("XSCREENSAVER_WINDOW");
    if (env != NULL) {
        window_id = strtol(env, (char **) NULL, 0); /* Base 0 autodetects hex/dec */
	}
    if (window_id != 0) {
        Display *display = NULL;
        XWindowAttributes windowAttributes;
		if ((display = XOpenDisplay(NULL)) != NULL) { /* Use the default display */
			XGetWindowAttributes(display, (Window) window_id, &windowAttributes);
			XCloseDisplay(display);
			width = windowAttributes.width;
			height = windowAttributes.height;
		}
    }
    return 0;
}

SDL_Window * newScreen(char *title, int h, int v) {
    if (window_id != 0) {
        return SDL_CreateWindowFrom((const void *)window_id);
    } else {
        return SDL_CreateWindow(title, SDL_WINDOWPOS_CENTERED, SDL_WINDOWPOS_CENTERED, h, v, 
            SDL_WINDOW_FULLSCREEN_DESKTOP  | SDL_WINDOW_SHOWN | SDL_WINDOW_BORDERLESS | SDL_WINDOW_ALWAYS_ON_TOP | SDL_WINDOW_SKIP_TASKBAR | SDL_WINDOW_TOOLTIP
        );
    }
}

SDL_Renderer * newRenderer( SDL_Window * screen ) {
    SDL_Renderer * r = SDL_CreateRenderer(screen, -1, SDL_RENDERER_PRESENTVSYNC | SDL_RENDERER_ACCELERATED); // SDL_RENDERER_SOFTWARE ); // SDL_RENDERER_ACCELERATED  );
	return r;
}

void setScaleQuality(int n) {
	switch(n) {
	case 1:
		SDL_SetHint(SDL_HINT_RENDER_SCALE_QUALITY, "1");
		break;
	case 2:
		SDL_SetHint(SDL_HINT_RENDER_SCALE_QUALITY, "2");
	}
}

int isNull(void *pointer) {
	if (pointer == NULL) {
		return 1;
	} else {
		return 0;
	}
}
