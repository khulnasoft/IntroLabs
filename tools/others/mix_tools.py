# coding=utf-8
from core import IntroLabs
from core import IntroLabsCollection


class TerminalMultiplexer(IntroLabs):
    TITLE = "Terminal Multiplexer"
    DESCRIPTION = "Terminal Multiplexer is a tiling terminal emulator that " \
                  "allows us to open \n several terminal sessions inside one " \
                  "single window."
    INSTALL_COMMANDS = ["sudo apt-get install tilix"]

    def __init__(self):
        super(TerminalMultiplexer, self).__init__(runnable = False)


class MixTools(IntroLabsCollection):
    TITLE = "Mix tools"
    TOOLS = [TerminalMultiplexer()]
