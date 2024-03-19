# coding=utf-8
from core import IntroLabs
from core import IntroLabsCollection


class GoSpider(IntroLabs):
    TITLE = "Gospider"
    DESCRIPTION = "Gospider - Fast web spider written in Go"
    INSTALL_COMMANDS = ["sudo go get -u github.com/jaeles-project/gospider"]
    PROJECT_URL = "https://github.com/jaeles-project/gospider"

    def __init__(self):
        super(GoSpider, self).__init__(runnable = False)


class WebCrawlingTools(IntroLabsCollection):
    TITLE = "Web crawling"
    TOOLS = [GoSpider()]
