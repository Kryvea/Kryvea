from dataclasses import asdict

from requests import Session


class Base:
    base_url: str
    session: Session

    id: str = ""
    created_at: str = ""
    updated_at: str = ""

    def dict(self) -> dict:
        return asdict(self)
