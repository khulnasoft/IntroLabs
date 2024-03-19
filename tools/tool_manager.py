# coding=utf-8
import os
import sys
from time import sleep

from core import IntroLabs
from core import IntroLabsCollection


class UpdateTool(IntroLabs):
    TITLE = "Update Tool or System"
    DESCRIPTION = "Update Tool or System"

    def __init__(self):
        super(UpdateTool, self).__init__([
            ("Update System", self.update_sys),
            ("Update Introlabs", self.update_ht)
        ], installable = False, runnable = False)

    def update_sys(self):
        os.system("sudo apt update && sudo apt full-upgrade -y")
        os.system(
            "sudo apt-get install tor openssl curl && sudo apt-get update tor openssl curl")
        os.system("sudo apt-get install python3-pip")

    def update_ht(self):
        os.system("sudo chmod +x /etc/;"
                  "sudo chmod +x /usr/share/doc;"
                  "sudo rm -rf /usr/share/doc/introlabs/;"
                  "cd /etc/;"
                  "sudo rm -rf /etc/introlabs/;"
                  "mkdir introlabs;"
                  "cd introlabs;"
                  "git clone https://github.com/Z4nzu/introlabs.git;"
                  "cd introlabs;"
                  "sudo chmod +x install.sh;"
                  "./install.sh")


class UninstallTool(IntroLabs):
    TITLE = "Uninstall IntroLabs"
    DESCRIPTION = "Uninstall IntroLabs"

    def __init__(self):
        super(UninstallTool, self).__init__([
            ('Uninstall', self.uninstall)
        ], installable = False, runnable = False)

    def uninstall(self):
        print("introlabs started to uninstall..\n")
        sleep(1)
        os.system("sudo chmod +x /etc/;"
                  "sudo chmod +x /usr/share/doc;"
                  "sudo rm -rf /usr/share/doc/introlabs/;"
                  "cd /etc/;"
                  "sudo rm -rf /etc/introlabs/;")
        print("\nIntrolabs Successfully Uninstalled... Goodbye.")
        sys.exit()


class ToolManager(IntroLabsCollection):
    TITLE = "Update or Uninstall | Introlabs"
    TOOLS = [
        UpdateTool(),
        UninstallTool()
    ]
