import json
from dataclasses import dataclass, field
from typing import Tuple

import utils.utils as utils
from models.base import Base


@dataclass
class Customer(Base):
    name: str = field(default_factory=utils.rand_company)
    language: str = field(default_factory=utils.rand_language)

    def getAll(self) -> list:
        response = self.session.get(self.base_url + "/customers")
        return response.json()

    def add(self) -> Tuple[str, str]:
        data = {
            "name": self.name,
            "language": self.language,
        }

        image = utils.rand_image()
        imageBytes = open(image, "rb").read()

        files = {
            "data": (None, json.dumps(data), "application/json"),
            "file": ("customer_logo.jpg", imageBytes, "image/jpeg"),
        }

        response = self.session.post(self.base_url + "/admin/customers", files=files)
        jr = response.json()
        if response.status_code == 201:
            self.id = jr.get("customer_id")
            return self.id, ""
        return "", jr.get("error")

    def getAssessments(self) -> list:
        response = self.session.get(self.base_url + f"/customers/{self.id}/assessments")
        return response.json()

    def getTargets(self) -> list:
        response = self.session.get(self.base_url + f"/customers/{self.id}/targets")
        return response.json()
