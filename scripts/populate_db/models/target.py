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
    tag: str = field(default_factory=utils.rand_target_tag)

    def add(self) -> Tuple[str, str]:
        data = {
            "customer_id": self.customer_id,
            "upsert": [
                {
                    "ipv4": self.ipv4,
                    "ipv6": self.ipv6,
                    "fqdn": self.fqdn,
                    "tag": self.tag,
                    "port": self.port,
                    "protocol": self.protocol,
                }
            ],
            "delete": [],
        }
        response = self.session.patch(self.base_url + "/targets/bulk", json=data)
        json_response = response.json()
        if response.status_code == 200:
            inserted = json_response.get("inserted_ids") or []
            if not inserted:
                return "", json_response.get("error") or "no inserted id returned"
            self.id = inserted[0]
            return self.id, ""
        return "", json_response.get("error")
