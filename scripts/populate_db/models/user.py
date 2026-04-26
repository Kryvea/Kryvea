from dataclasses import dataclass, field
from typing import List, Tuple

import utils.utils as utils
from models.assessment import Assessment
from models.base import Base
from models.customer import Customer


@dataclass
class User(Base):
    username: str = field(default_factory=utils.rand_username)
    password: str = field(default="Kryveapassword1!")
    disabled_at: str = field(default="")
    role: str = field(default=utils.ROLE_USER)
    customers: List[Customer] = field(default_factory=list)
    assessments: List[Assessment] = field(default_factory=list)

    def add(self) -> Tuple[str, str]:
        data = {
            "username": self.username,
            "password": self.password,
            "role": self.role,
        }
        response = self.session.post(self.base_url + "/admin/users", json=data)
        json_response = response.json()
        if response.status_code == 201:
            self.id = json_response.get("user_id")
            return self.id, ""
        return "", json_response.get("error")

    def login(self, first_login_password: str = "") -> bool:
        data = {
            "username": self.username,
            "password": self.password,
        }
        response = self.session.post(self.base_url + "/login", json=data)
        if response.status_code == 200:
            return True
        elif (
            response.status_code == 403
            and response.json().get("error") == "Password expired"
        ):
            if first_login_password == "":
                print(
                    f"User {self.username} has expired password and no first login password is set, cannot login"
                )
                return False
            else:
                print(
                    f"User {self.username} has expired password, resetting password to {first_login_password}"
                )
                if self.reset_password(first_login_password):
                    self.password = first_login_password
                    return True
                else:
                    print(f"Failed to reset password for user {self.username}")
                    return False
        return False

    def reset_password(self, new_password: str) -> bool:
        data = {
            "password": new_password,
        }
        response = self.session.post(self.base_url + "/password/reset", json=data)
        if response.status_code == 200:
            self.password = new_password
            return True
        return False
