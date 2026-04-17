import base64
import json
from dataclasses import dataclass, field
from typing import List, Tuple

import utils.utils as utils
from models.base import Base


@dataclass
class LineCol:
    line: int
    col: int


@dataclass
class HighlightedText:
    start: LineCol
    end: LineCol


@dataclass
class PocData(Base):
    index: int = 0
    type: str = field(default_factory=utils.rand_poc_type)
    description: str = field(default_factory=utils.rand_poc_description)
    uri: str = field(default_factory=utils.rand_uri)

    request: str = field(default="")
    request_highlight: List[HighlightedText] = field(default_factory=list)
    response: str = field(default="")
    response_highlight: List[HighlightedText] = field(default_factory=list)

    image_data: str = field(default="")
    image_reference: str = field(default="")
    image_caption: str = field(default="")

    text_language: str = field(default="")
    text_data: str = field(default="")
    text_highlight: List[HighlightedText] = field(default_factory=list)
    starting_line_number: int = field(default=1)

    def __post_init__(self):
        if self.type == utils.POC_TYPE_REQUEST:
            self.request = utils.rand_request()
            self.response = utils.rand_response()

            lines = self.request.splitlines()
            num_lines = len(lines)
            self.request_highlight = dict_to_highlighted_text(
                utils.rand_highlighted_text(
                    lines=lines,
                )
            )

            lines = self.response.splitlines()
            num_lines = len(lines)
            self.response_highlight = dict_to_highlighted_text(
                utils.rand_highlighted_text(
                    lines=lines,
                )
            )

        if self.type == utils.POC_TYPE_IMAGE:
            image = utils.rand_image()
            imageBytes = open(image, "rb").read()
            self.image_data = base64.b64encode(imageBytes).decode("utf-8")

            self.image_reference = f"image{self.index}"
            self.image_caption = utils.rand_caption()

        if self.type == utils.POC_TYPE_TEXT:
            self.text_language = utils.rand_code_language()
            self.text_data = utils.rand_code_snippet(language=self.text_language)


@dataclass
class Poc(Base):
    poc_data: List[PocData]
    vulnerability_id: str

    def add(self) -> Tuple[str, str]:
        data = [
            {
                "type": poc_data.type,
                "index": poc_data.index,
                "description": poc_data.description,
                "uri": poc_data.uri,
                "request": poc_data.request,
                "request_highlight": highlighted_text_to_dict(
                    poc_data.request_highlight
                ),
                "response": poc_data.response,
                "response_highlight": highlighted_text_to_dict(
                    poc_data.response_highlight
                ),
                "image_reference": poc_data.image_reference,
                "image_caption": poc_data.image_caption,
                "text_language": poc_data.text_language,
                "text_data": poc_data.text_data,
                "text_highlight": highlighted_text_to_dict(poc_data.text_highlight),
                "starting_line_number": poc_data.starting_line_number,
            }
            for poc_data in self.poc_data
        ]

        files = {
            "pocs": (None, json.dumps(data), "application/json"),
        }

        for poc_data in self.poc_data:
            if poc_data.type == utils.POC_TYPE_IMAGE and poc_data.image_data != "":
                files[f"image{poc_data.index}"] = (
                    "example.jpg",
                    base64.b64decode(poc_data.image_data),
                    "image/jpeg",
                )

        response = self.session.put(
            self.base_url + f"/vulnerabilities/{self.vulnerability_id}/pocs",
            files=files,
        )
        json_response = response.json()
        if response.status_code == 200:
            return "", ""
        return "", json_response.get("error")


def dict_to_highlighted_text(highlight_data: dict) -> List[HighlightedText]:
    highlights = []
    for item in highlight_data:
        start = item["start"]
        end = item["end"]
        highlights.append(
            HighlightedText(
                start=LineCol(line=start["line"], col=start["col"]),
                end=LineCol(line=end["line"], col=end["col"]),
            )
        )
    return highlights


def highlighted_text_to_dict(highlights: List[HighlightedText]) -> List[dict]:
    highlight_data = []
    for highlight in highlights:
        highlight_data.append(
            {
                "start": {"line": highlight.start.line, "col": highlight.start.col},
                "end": {"line": highlight.end.line, "col": highlight.end.col},
            }
        )
    return highlight_data
