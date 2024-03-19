# coding=utf-8
from core import IntroLabs
from core import IntroLabsCollection


class Stitch(IntroLabs):
    TITLE = "Stitch"
    DESCRIPTION = "Stitch is a cross platform python framework.\n" \
                  "which allows you to build custom payloads\n" \
                  "For Windows, Mac and Linux."
    INSTALL_COMMANDS = [
        "sudo git clone https://github.com/nathanlopez/Stitch.git",
        "cd Stitch;sudo pip install -r lnx_requirements.txt"
    ]
    RUN_COMMANDS = ["cd Stitch;python main.py"]
    PROJECT_URL = "https://github.com/nathanlopez/Stitch"


class Pyshell(IntroLabs):
    TITLE = "Pyshell"
    DESCRIPTION = "Pyshell is a Rat Tool that can be able to download & upload " \
                  "files,\n Execute OS Command and more.."
    INSTALL_COMMANDS = [
        "sudo git clone https://github.com/knassar702/Pyshell.git;"
        "sudo pip install pyscreenshot python-nmap requests"
    ]
    RUN_COMMANDS = ["cd Pyshell;./Pyshell"]
    PROJECT_URL = "https://github.com/knassar702/pyshell"


class RemoteAdministrationTools(IntroLabsCollection):
    TITLE = "Remote Administrator Tools (RAT)"
    TOOLS = [
        Stitch(),
        Pyshell()
    ]
