from dataclasses import dataclass, field
from typing import Dict, List, Tuple

import utils.utils as utils
from models.base import Base


@dataclass
class AssessmentType:
    short: str
    full: str


@dataclass
class Assessment(Base):
    customer_id: str
    targets: List[str]

    name: str = field(default_factory=utils.rand_assessment_name)
    language: str = field(default_factory=utils.rand_language)
    start_date_time: str = field(default_factory=utils.rand_date_decade)
    end_date_time: str = field(default_factory=utils.rand_date_future)
    cvss_versions: Dict[str, bool] = field(default_factory=utils.rand_cvss_versions)
    status: str = field(default_factory=utils.rand_status)
    type: AssessmentType = field(default=None)
    environment: str = field(default_factory=utils.rand_environment)
    testing_type: str = field(default_factory=utils.rand_testing_type)
    osstmm_vector: str = field(default_factory=utils.rand_osstmm_vector)

    vulnerability_count: int = 0
    is_owned: bool = False

    def __post_init__(self):
        if not self.type:
            short, full = utils.rand_assessment_type()
            self.type = AssessmentType(short=short, full=full)

    def add(self) -> Tuple[str, str]:
        data = {
            "customer_id": self.customer_id,
            "name": self.name,
            "language": self.language,
            "start_date_time": self.start_date_time,
            "end_date_time": self.end_date_time,
            "cvss_versions": self.cvss_versions,
            "targets": self.targets,
            "status": self.status,
            "type": {
                "short": self.type.short,
                "full": self.type.full,
            },
            "environment": self.environment,
            "testing_type": self.testing_type,
            "osstmm_vector": self.osstmm_vector,
        }
        response = self.session.post(self.base_url + "/assessments", json=data)
        json_response = response.json()
        if response.status_code == 201:
            self.id = json_response.get("assessment_id")
            return self.id, ""
        return "", json_response.get("error")
