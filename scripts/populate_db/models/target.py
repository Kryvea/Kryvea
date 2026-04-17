from dataclasses import dataclass, field
from typing import Tuple

import utils.utils as utils
from models.base import Base


@dataclass
class Target(Base):
    customer_id: str

    fqdn: str = field(default_factory=utils.rand_hostname)
    ipv4: str = field(default_factory=utils.rand_ipv4)
    ipv6: str = field(default_factory=utils.rand_ipv6)
    port: int = field(default_factory=utils.rand_port)
    protocol: str = field(default_factory=utils.rand_protocol)
    name: str = field(default_factory=utils.rand_target_name)

    def add(self) -> Tuple[str, str]:
        data = {
            "fqdn": self.fqdn,
            "customer_id": self.customer_id,
            "ipv4": self.ipv4,
            "ipv6": self.ipv6,
            "port": self.port,
            "protocol": self.protocol,
            "name": self.name,
        }
        response = self.session.post(self.base_url + "/targets", json=data)
        json_response = response.json()
        if response.status_code == 201:
            self.id = json_response.get("target_id")
            return self.id, ""
        return "", json_response.get("error")
