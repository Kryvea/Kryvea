import json
import os
import random
import string
import urllib.parse
from datetime import datetime, timedelta
from typing import Dict, List

import requests
from faker import Faker

fake = Faker()


def rand_string(length: int = 8) -> str:
    return "".join(random.choices(string.ascii_letters, k=length))


def rand_company() -> str:
    return fake.company()


def rand_description() -> str:
    return fake.text(max_nb_chars=200)


def rand_username() -> str:
    return fake.user_name()


def rand_cvss_versions() -> Dict[str, bool]:
    versions = ["3.1", "4.0"]
    result = {version: random.choice([True, False]) for version in versions}

    if not any(result.values()):
        result[random.choice(versions)] = True

    return result


def rand_cvss(version: str) -> str:
    if version == "3.1":
        vectors = [
            "CVSS:3.1/AV:A/AC:H/PR:L/UI:N/S:C/C:L/I:L/A:L",
            "CVSS:3.1/AV:N/AC:H/PR:N/UI:R/S:U/C:N/I:L/A:H",
            "CVSS:3.1/AV:N/AC:H/PR:L/UI:R/S:C/C:L/I:L/A:H",
            "CVSS:3.1/AV:N/AC:H/PR:L/UI:R/S:C/C:L/I:H/A:N",
            "CVSS:3.1/AV:N/AC:H/PR:L/UI:N/S:C/C:L/I:L/A:N",
            "CVSS:3.1/AV:L/AC:H/PR:L/UI:N/S:U/C:N/I:L/A:N",
            "CVSS:3.1/AV:A/AC:H/PR:H/UI:N/S:U/C:N/I:H/A:N",
            "CVSS:3.1/AV:P/AC:H/PR:H/UI:N/S:C/C:N/I:L/A:N",
            "CVSS:3.1/AV:N/AC:L/PR:L/UI:R/S:C/C:L/I:L/A:N",
            "CVSS:3.1/AV:A/AC:H/PR:L/UI:N/S:C/C:H/I:H/A:H",
        ]
        return random.choice(vectors)
    vectors = [
        "CVSS:4.0/AV:A/AC:H/AT:P/PR:L/UI:P/VC:L/VI:L/VA:L/SC:L/SI:L/SA:L",
        "CVSS:4.0/AV:A/AC:H/AT:N/PR:L/UI:P/VC:H/VI:H/VA:L/SC:L/SI:L/SA:L",
        "CVSS:4.0/AV:A/AC:H/AT:N/PR:L/UI:A/VC:H/VI:N/VA:L/SC:L/SI:H/SA:L",
        "CVSS:4.0/AV:N/AC:L/AT:N/PR:L/UI:A/VC:H/VI:N/VA:L/SC:L/SI:H/SA:L",
        "CVSS:4.0/AV:N/AC:L/AT:P/PR:L/UI:P/VC:H/VI:N/VA:L/SC:L/SI:H/SA:L",
        "CVSS:4.0/AV:N/AC:L/AT:P/PR:L/UI:P/VC:H/VI:L/VA:L/SC:N/SI:H/SA:H",
    ]
    return random.choice(vectors)


def rand_vulnerability_description() -> str:
    descriptions = [
        "This vulnerability allows an attacker to execute arbitrary code on the target system.",
        "An information disclosure vulnerability that exposes sensitive data to unauthorized users.",
        "A denial of service vulnerability that can crash the application or make it unresponsive.",
        "An authentication bypass vulnerability that allows attackers to gain unauthorized access.",
        "A cross-site scripting (XSS) vulnerability that can be exploited to inject malicious scripts.",
    ]
    return random.choice(descriptions)


def rand_vulnerability_remediation() -> str:
    remediations = [
        "Update the software to the latest version to patch the vulnerability.",
        "Implement input validation to prevent injection attacks.",
        "Configure proper access controls to restrict unauthorized access.",
        "Use secure coding practices to mitigate XSS vulnerabilities.",
        "Regularly audit and monitor systems for potential vulnerabilities.",
    ]
    return random.choice(remediations)


def rand_hostname() -> str:
    return fake.hostname()


def rand_port() -> int:
    ports = [
        20,
        21,
        22,
        23,
        25,
        53,
        67,
        68,
        69,
        80,
        110,
        111,
        123,
        135,
        137,
        138,
        139,
        143,
        161,
        162,
        179,
        389,
        443,
        445,
        465,
        514,
        515,
        587,
        631,
        636,
        993,
        995,
        1080,
        1433,
        1434,
        1723,
        1812,
        1813,
        2049,
        2181,
        3306,
        3389,
        3690,
        4444,
        5060,
        5432,
        5900,
        6379,
        8080,
        8443,
    ]
    return random.choice(ports)


def rand_ipv4() -> str:
    return fake.ipv4()


def rand_ipv6() -> str:
    return fake.ipv6()


def rand_protocol() -> str:
    protocols = [
        "tcp",
        "udp",
        "http",
        "https",
        "ftp",
        "ssh",
        "dns",
        "smtp",
        "imap",
        "pop3",
        "snmp",
        "rdp",
    ]
    return random.choice(protocols)


def rand_target_name() -> str:
    names = [
        "Android",
        "iOS",
        "Backend",
        "Api",
    ]
    return random.choice(names)


def rand_name(n=1) -> str:
    names = [
        "Ace",
        "Blaze",
        "Nova",
        "Zane",
        "Kai",
        "Orion",
        "Jett",
        "Echo",
        "Maverick",
        "Axel",
        "Ryder",
        "Phoenix",
        "Storm",
        "Dash",
        "Sable",
        "Ember",
        "Zephyr",
        "Titan",
        "Knox",
        "Luna",
        "Indigo",
        "Raven",
        "Aspen",
        "Atlas",
        "Juno",
        "Onyx",
        "Sage",
        "Vega",
        "Zara",
        "Xander",
        "Aria",
        "Dante",
        "Hunter",
        "Skye",
        "Rogue",
        "Kairos",
        "Hawk",
        "Shadow",
        "Nyx",
        "Lyric",
    ]
    return " ".join(random.choices(names, k=n))


def rand_language() -> str:
    languages = ["it", "en", "fr", "de", "es"]
    return random.choice(languages)


categories_dict = {
    "A01:2021": "Broken Access Control",
    "A02:2021": "Cryptographic Failures",
    "A03:2021": "Injection",
    "A04:2021": "Insecure Design",
    "A05:2021": "Security Misconfiguration",
    "A06:2021": "Vulnerable and Outdated Components",
    "A07:2021": "Identification and Authentication Failures",
    "A08:2021": "Software and Data Integrity Failures",
    "A09:2021": "Security Logging and Monitoring Failures",
    "A10:2021": "Server-Side Request Forgery",
    "KRYVEA-01": "SQL Injection",
    "KRYVEA-02": "Cross-Site Scripting",
    "KRYVEA-03": "Buffer Overflow",
    "KRYVEA-04": "Command Injection",
    "KRYVEA-05": "Path Traversal",
    "KRYVEA-06": "Remote Code Execution",
    "KRYVEA-07": "Denial of Service",
    "KRYVEA-08": "Information Disclosure",
    "KRYVEA-09": "Privilege Escalation",
    "KRYVEA-10": "Cross-Site Request Forgery",
    "KRYVEA-11": "XML External Entity",
    "KRYVEA-12": "Insecure Direct Object Reference",
    "CWE-79": "Improper Neutralization of Input During Web Page Generation (Cross-site Scripting)",
    "CWE-89": "Improper Neutralization of Special Elements used in an SQL Command (SQL Injection)",
    "CWE-120": "Buffer Copy without Checking Size of Input (Classic Buffer Overflow)",
    "CWE-125": "Out-of-bounds Read",
    "CWE-20": "Improper Input Validation",
    "CWE-200": "Exposure of Sensitive Information to an Unauthorized Actor",
    "CWE-287": "Improper Authentication",
    "CWE-22": "Improper Limitation of a Pathname to a Restricted Directory (Path Traversal)",
    "CWE-352": "Cross-Site Request Forgery (CSRF)",
    "CWE-476": "NULL Pointer Dereference",
}


def rand_category_index() -> str:
    return f"{random.choice(list(categories_dict.keys()))}"


def rand_category_identifiers(n: int) -> list:
    return random.sample(list(categories_dict.keys()), k=n)


def rand_category_name(index: str = "") -> str:
    if index and index in categories_dict:
        return categories_dict[index] + f" {rand_string(4)}"
    return random.choice(list(categories_dict.values())) + f" {rand_string(4)}"


def rand_category_subname() -> str:
    subnames = [
        "Business Logic",
        "SQL Injection",
        "Cross-Site Scripting",
        "Authentication Bypass",
        "Insecure Direct Object Reference",
        "XXE",
    ]
    return random.choice(subnames)


def rand_generic_remediation() -> dict:
    remediation = {
        "en": "Use strong, centralized server-side access controls based on user roles and context. Avoid relying on client-side checks. Validate permissions on all endpoints and adopt a deny-by-default approach. Regularly audit controls and test for bypass methods like URL tampering. Implement RBAC or ABAC frameworks where possible.",
        "it": "Usare controlli di accesso sicuri e centralizzati lato server basati su ruoli e contesto. Evitare controlli solo lato client. Validare permessi su tutti gli endpoint e adottare una strategia di negazione predefinita. Eseguire audit regolari e testare bypass come manipolazione URL. Integrare RBAC o ABAC quando possibile.",
        "fr": "Utilisez des contrôles d'accès centralisés et robustes côté serveur basés sur les rôles et le contexte. Ne comptez pas uniquement sur le client. Validez les permissions sur tous les points d'accès et adoptez une stratégie de refus par défaut. Auditez régulièrement et testez les contournements comme la manipulation d'URL. Implémentez RBAC ou ABAC si possible.",
        "de": "Verwenden Sie starke, zentralisierte serverseitige Zugriffskontrollen basierend auf Rollen und Kontext. Verlassen Sie sich nicht nur auf Client-seitige Prüfungen. Validieren Sie Berechtigungen an allen Endpunkten und verwenden Sie eine Verweigerung-als-Standard-Strategie. Prüfen Sie regelmäßig und testen Sie Umgehungen wie URL-Manipulation. Implementieren Sie RBAC oder ABAC, wenn möglich.",
        "es": "Use controles de acceso robustos y centralizados en el servidor basados en roles y contexto. No confíe solo en controles del cliente. Valide permisos en todos los puntos finales y adopte una política de denegación por defecto. Audite regularmente y pruebe técnicas de elusión como manipulación de URL. Implemente RBAC o ABAC cuando sea posible.",
    }

    return remediation


def rand_generic_description() -> dict:
    description = {
        "en": "Broken Access Control occurs when an application fails to properly enforce restrictions on what authenticated users are allowed to do. This includes bypassing authorization checks, accessing data or functions beyond their intended permissions, or escalating privileges. Common manifestations include Insecure Direct Object References (IDOR), missing or flawed role-based access controls, forced browsing to unauthorized endpoints, and privilege escalation via parameter manipulation. Exploiting these weaknesses can lead to unauthorized access to sensitive data, account takeover, or full system compromise.",
        "it": "Una vulnerabilità di tipo Broken Access Control si verifica quando un'applicazione non riesce a far rispettare correttamente le restrizioni su ciò che gli utenti autenticati sono autorizzati a fare. Ciò include il bypass dei controlli di autorizzazione, l'accesso a dati o funzioni oltre i permessi previsti o l'escalation dei privilegi. Le manifestazioni comuni includono IDOR (Insecure Direct Object References), controlli di accesso basati sui ruoli assenti o difettosi, accesso forzato a endpoint non autorizzati ed escalation dei privilegi tramite manipolazione dei parametri. Lo sfruttamento di queste vulnerabilità può portare ad accessi non autorizzati a dati sensibili, al controllo degli account o al compromesso completo del sistema.",
        "fr": "Les vulnérabilités de type Broken Access Control se produisent lorsqu'une application n'applique pas correctement les restrictions sur ce que les utilisateurs authentifiés sont autorisés à faire. Cela inclut le contournement des vérifications d'autorisation, l'accès à des données ou fonctionnalités non autorisées, ou encore l'élévation de privilèges. Parmi les manifestations courantes : IDOR (Insecure Direct Object References), absence ou mauvaise mise en œuvre du contrôle d'accès basé sur les rôles, accès forcé à des endpoints protégés, et manipulation des paramètres pour obtenir plus de privilèges. L'exploitation de ces failles peut entraîner un accès non autorisé à des données sensibles, la prise de contrôle de comptes ou le compromis complet du système.",
        "de": "Broken Access Control tritt auf, wenn eine Anwendung nicht ordnungsgemäß durchsetzt, welche Aktionen authentifizierte Benutzer ausführen dürfen. Dazu gehört das Umgehen von Autorisierungsprüfungen, der Zugriff auf Daten oder Funktionen über die vorgesehenen Berechtigungen hinaus oder die Eskalation von Rechten. Häufige Manifestationen sind Insecure Direct Object References (IDOR), fehlende oder fehlerhafte rollenbasierte Zugriffskontrollen, erzwungenes Browsen zu unautorisierten Endpunkten und Rechteeskalation durch Parameter-Manipulation. Das Ausnutzen dieser Schwächen kann zu unautorisiertem Zugriff auf sensible Daten, Kontoübernahme oder vollständiger Systemkompromittierung führen.",
        "es": "Las vulnerabilidades de tipo Broken Access Control se producen cuando una aplicación no aplica correctamente las restricciones sobre lo que los usuarios autenticados pueden hacer. Esto incluye eludir las comprobaciones de autorización, acceder a datos o funciones más allá de sus permisos previstos o escalar privilegios. Las manifestaciones comunes incluyen IDOR (Insecure Direct Object References), controles de acceso basados en roles ausentes o defectuosos, navegación forzada a puntos finales no autorizados y escalada de privilegios a través de la manipulación de parámetros. La explotación de estas debilidades puede llevar a un acceso no autorizado a datos sensibles, toma de control de cuentas o compromiso total del sistema.",
    }
    return description


def rand_vulnerability_title() -> str:
    titles = [
        "Unauthorized Access to User Data",
        "Privilege Escalation via Parameter Manipulation",
        "Insecure Direct Object Reference in User Profile",
        "Cross-Site Scripting in User Comments",
        "SQL Injection in Search Functionality",
        "Remote Code Execution via File Upload",
        "Denial of Service via Resource Exhaustion",
        "Information Disclosure through Error Messages",
        "Cross-Site Request Forgery on Sensitive Actions",
    ]
    return random.choice(titles)


def rand_vulnerability_remediation() -> str:
    remediations = [
        "Ensure proper access controls are implemented on all endpoints.",
        "Validate user permissions server-side before processing requests.",
        "Implement role-based access control (RBAC) to restrict actions based on user roles.",
        "Regularly audit access controls and test for vulnerabilities like IDOR.",
        "Educate developers on secure coding practices to prevent access control flaws.",
    ]
    return random.choice(remediations)


def rand_vulnerability_description() -> str:
    descriptions = [
        "The endpoint GET /api/orders/{orderId} allows users to retrieve order details by order ID. During testing, it was found that the application does not validate whether the authenticated user owns the order. By incrementing the orderId parameter, I was able to access multiple other users' order data, including names, addresses, and product information. For example, accessing /api/orders/1021 while authenticated as user A returned order data belonging to user B.",
        "While exploring hidden endpoints, I discovered that the admin panel route GET /admin/user/list is accessible to any logged-in user without performing any role-based access checks. When accessed by a non-admin account, the endpoint returned a complete list of registered users, including their email addresses and roles. This exposes sensitive user data and could allow further privilege abuse or social engineering.",
        'The PUT /api/user/profile endpoint allows users to update their profile information. During testing, I inserted an additional parameter into the JSON payload: "role": "admin". The server failed to restrict or validate role changes, and the request succeeded. The modified user was then able to access admin-only endpoints like /admin/settings and /admin/users, demonstrating a clear vertical privilege escalation flaw.',
    ]
    return random.choice(descriptions)


def rand_vulnerability_status() -> str:
    statuses = [
        "Open",
        "In Progress",
        "Closed",
    ]
    return random.choice(statuses)


def rand_status() -> str:
    statuses = [
        "On Hold",
        "In Progress",
        "Completed",
    ]
    return random.choice(statuses)


def rand_assessment_name() -> str:
    prefixes = [
        "Next-Gen",
        "Unlocking",
        "Redefining",
        "Seamless",
        "Intelligent",
        "Optimizing",
        "Harnessing",
        "Decoding",
        "Exploring",
        "Mastering",
        "Advancing",
        "Elevating",
        "Transforming",
    ]
    return f"{random.choice(prefixes)} {rand_name()}"


def rand_assessment_type() -> tuple[str, str]:
    typesKey = ["WAPT", "VAPT", "MAPT", "IoT", "Red Team Assessment"]
    typesValue = [
        "Web Application Penetration Test",
        "Vulnerability Assessment and Penetration Testing",
        "Mobile Application Penetration Test",
        "Internet of Things Penetration Test",
        "Red Team Assessment",
    ]
    selected = random.randint(0, len(typesKey) - 1)
    return typesKey[selected], typesValue[selected]


def rand_date_decade() -> str:
    return fake.date_time_this_decade().isoformat() + "Z"


def rand_date_future() -> str:
    now = datetime.now()
    future_date = now + timedelta(days=random.randint(1, 365))
    return future_date.isoformat() + "Z"


def rand_environment() -> str:
    environments = ["Pre-Production", "Production"]
    return random.choice(environments)


def rand_testing_type() -> str:
    types = ["Black Box", "White Box", "Grey Box"]
    return random.choice(types)


def rand_osstmm_vector() -> str:
    vectors = [
        "Inside to Inside",
        "Inside to Outside",
        "Outside to Outside",
        "Outside to Inside",
    ]
    return random.choice(vectors)


def rand_urls() -> list:
    return [fake.url() for _ in range(random.randint(1, 4))]


def rand_uri() -> str:
    return fake.uri()


def rand_source() -> str:
    sources = [
        "nessus",
        "burp",
        "owasp_api",
        "owasp_mobile",
        "owasp_web",
        "cwe",
        "capec",
    ]
    return random.choice(sources)


ROLE_ADMIN = "admin"
ROLE_USER = "user"

POC_TYPE_TEXT = "text"
POC_TYPE_REQUEST = "request/response"
POC_TYPE_IMAGE = "image"
POC_TYPES = [POC_TYPE_TEXT, POC_TYPE_REQUEST, POC_TYPE_IMAGE]


def rand_poc_type() -> str:
    return random.choice(POC_TYPES)


code_snippets = {
    "python": """import random

def roll_dice():
    return random.randint(1, 6)

print(f"You rolled a {roll_dice()}")
""",
    "javascript": """function greet(name) {
  console.log(`Hello, ${name}!`);
}

greet("Alice");
""",
    "php": """<?php
function square($n) {
    return $n * $n;
}
echo square(5);
?>
""",
    "java": """public class HelloWorld {
    public static void main(String[] args) {
        System.out.println("Hello, world!");
    }
}
""",
    "c": """#include <stdio.h>

int main() {
    printf("Welcome to C programming!\n");
    return 0;
}
""",
    "c++": """#include <iostream>

int main() {
    std::cout << "C++ is powerful!" << std::endl;
    return 0;
}
""",
    "csharp": """using System;

class Program {
    static void Main() {
        Console.WriteLine("Hello from C#!");
    }
}
""",
    "ruby": """def multiply(a, b)
  a * b
end

puts multiply(3, 7)
""",
    "go": """package main
import "fmt"

func main() {
    fmt.Println("Hello from Go!")
}
""",
    "rust": """fn main() {
    println!("Rust is safe and fast!");
}
""",
}


def rand_code_language() -> str:
    return random.choice(list(code_snippets.keys()))


def rand_code_snippet(language: str = "") -> str:
    if language and language in code_snippets:
        return code_snippets[language]
    return random.choice(list(code_snippets.values()))


def rand_poc_description() -> str:
    descriptions = [
        "This POC demonstrates a vulnerability in the application that allows for arbitrary code execution.",
        "This POC shows how to exploit an SQL injection vulnerability to extract sensitive data from the database.",
        "This POC illustrates a cross-site scripting (XSS) attack that can be used to steal user cookies.",
        "This POC highlights a denial of service vulnerability that can crash the application server.",
        "This POC provides an example of how to bypass authentication mechanisms using session fixation.",
    ]
    return random.choice(descriptions)


def rand_request() -> str:
    paths = [
        f"/products/{fake.random_int(min=1, max=9999)}/details",
        f"/user/{fake.user_name()}/profile",
        f"/blog/{fake.word()}/{fake.random_int(min=100, max=999)}",
        f"/api/v1/{fake.word()}/{fake.random_number(digits=5)}",
    ]
    path = random.choice(paths)
    host = fake.hostname()
    headers = {
        "User-Agent": fake.user_agent(),
        "Accept": random.choice(
            [
                "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
                "application/json, text/plain, */*",
            ]
        ),
        "Accept-Language": random.choice(
            ["en-US,en;q=0.9", "fr-FR,fr;q=0.8,en-US;q=0.5,en;q=0.3"]
        ),
        "Accept-Encoding": random.choice(["gzip, deflate, br", "compress, gzip"]),
        "Connection": random.choice(["keep-alive", "close"]),
        "Referer": fake.url(),
        "Authorization": f"Bearer {fake.sha256()}",
        "X-Request-ID": fake.uuid4(),
        "Cache-Control": random.choice(["no-cache", "max-age=0"]),
    }
    content_type = random.choice(
        ["application/json", "application/x-www-form-urlencoded"]
    )
    json_data = {
        "name": fake.name(),
        "email": fake.email(),
        "address": {
            "street": fake.street_address(),
            "city": fake.city(),
            "zipcode": fake.zipcode(),
        },
        "phone_numbers": [fake.phone_number() for _ in range(3)],
        "age": fake.random_int(min=18, max=90),
        "active": fake.boolean(),
        "created_at": fake.iso8601(),
    }
    encoded_form_data = urllib.parse.urlencode(
        {
            "name": json_data["name"],
            "email": json_data["email"],
            "age": json_data["age"],
            "active": json_data["active"],
        }
    )
    body = (
        encoded_form_data
        if content_type == "application/x-www-form-urlencoded"
        else json.dumps(json_data, indent=2)
    )
    request = f"""POST {path} HTTP/1.1
Host: {host}
Content-Type: {content_type}
Content-Length: {len(body.encode('utf-8'))}
{chr(10).join(f"{k}: {v}" for k, v in headers.items())}

{body}
"""
    return request


def rand_response() -> str:
    status_codes = {
        200: "OK",
        201: "Created",
        202: "Accepted",
        204: "No Content",
        301: "Moved Permanently",
        302: "Found",
        400: "Bad Request",
        401: "Unauthorized",
        403: "Forbidden",
        404: "Not Found",
        409: "Conflict",
        500: "Internal Server Error",
        503: "Service Unavailable",
    }

    status_code = random.choice(list(status_codes.keys()))
    status_message = status_codes[status_code]

    response_body = {
        "id": fake.uuid4(),
        "status": status_message,
        "message": fake.sentence(nb_words=6),
        "timestamp": datetime.utcnow().isoformat() + "Z",
        "data": (
            {
                "name": fake.name(),
                "email": fake.email(),
                "registered": fake.boolean(),
                "profile_views": fake.random_int(min=0, max=10000),
            }
            if status_code in (200, 201)
            else None
        ),
        "error": (
            {"code": status_code, "description": status_message}
            if status_code >= 400
            else None
        ),
    }

    response_body = {k: v for k, v in response_body.items() if v is not None}
    json_body = json.dumps(response_body, indent=2)

    headers = {
        "Content-Type": "application/json; charset=utf-8",
        "Content-Length": str(len(json_body.encode("utf-8"))),
        "Cache-Control": random.choice(["no-store", "private", "public, max-age=3600"]),
        "X-Request-ID": fake.uuid4(),
        "Date": datetime.utcnow().strftime("%a, %d %b %Y %H:%M:%S GMT"),
        "Server": fake.user_agent(),
    }

    response = f"""HTTP/1.1 {status_code} {status_message}
{chr(10).join(f"{k}: {v}" for k, v in headers.items())}

{json_body}
"""

    return response


def rand_highlighted_text(lines: List[str] = []) -> List[dict]:
    if not lines:
        return []

    num_highlights = random.randint(1, min(3, len(lines)))
    highlights = []

    for _ in range(num_highlights):
        # Skip empty lines
        non_empty_lines = [i for i, l in enumerate(lines) if len(l) > 1]
        if not non_empty_lines:
            break

        line_index = random.choice(non_empty_lines)
        line = lines[line_index]
        line_len = len(line)

        # Ensure there's enough room to pick start and end
        if line_len < 2:
            continue

        start_col = random.randint(1, line_len - 1)
        end_col = (
            random.randint(start_col + 1, line_len)
            if start_col + 1 <= line_len
            else start_col + 1
        )

        highlights.append(
            {
                "start": {"line": line_index + 1, "col": start_col},
                "end": {"line": line_index + 1, "col": end_col},
            }
        )

    return highlights


def rand_image() -> str:
    directory = os.path.join(os.path.dirname(__file__), "..", "images")
    os.makedirs(directory, exist_ok=True)
    only_files = [
        f for f in os.listdir(directory) if os.path.isfile(os.path.join(directory, f))
    ]
    if not only_files:
        image_urls = [
            "https://avatars.githubusercontent.com/u/57672074",
            "https://avatars.githubusercontent.com/u/33416357",
            "https://avatars.githubusercontent.com/u/66215334",
            "https://stickerrs.com/wp-content/uploads/2024/03/Cat-Meme-Stickers-Featured.png",
            "https://i.pinimg.com/736x/b2/60/94/b26094970505bcd59c2e5fe8b6f41cf0.jpg",
        ]
        for url in image_urls:
            try:
                response = requests.get(url)
                if response.status_code == 200:
                    filename = os.path.join(directory, url.split("/")[-1])
                    if not filename.endswith(".png") and not filename.endswith(".jpg"):
                        filename += ".jpg"
                    with open(filename, "wb") as f:
                        f.write(response.content)
                    only_files.append(filename)
            except requests.RequestException as e:
                print(f"Failed to download image from {url}: {e}")

    return os.path.join(directory, random.choice(only_files)) if only_files else ""


def rand_caption() -> str:
    captions = [
        "Successful unauthorized access to user data via IDOR vulnerability.",
        "Admin panel accessed without proper authentication checks.",
        "Privilege escalation achieved by modifying role parameter in user profile update.",
        "SQL Injection allows dumping of the entire users table.",
        "Cross-site scripting payload executed on the comments section.",
        "Sensitive file downloaded via directory traversal exploit.",
        "Remote code execution achieved through unsafe deserialization.",
        "CSRF attack triggered an unintended password change request.",
        "Buffer overflow caused application crash during input fuzzing.",
        "Server-side request forgery used to scan internal network services.",
    ]
    return random.choice(captions)


class bcolors:
    OKBLUE = "\033[94m"
    OKGREEN = "\033[92m"
    WARNING = "\033[93m"
    FAIL = "\033[91m"
    ENDC = "\033[0m"
    BOLD = "\033[1m"
    UNDERLINE = "\033[4m"


def print_success(message: str) -> None:
    print(f"{bcolors.OKGREEN}[K] {message}{bcolors.ENDC}")


def print_error(message: str) -> None:
    print(f"{bcolors.FAIL}[E] {message}{bcolors.ENDC}")


def print_warning(message: str) -> None:
    print(f"{bcolors.WARNING}[W] {message}{bcolors.ENDC}")


def print_info(message: str) -> None:
    print(f"{bcolors.OKBLUE}[K] {message}{bcolors.ENDC}")


def print_header(message: str) -> None:
    print(f"{bcolors.HEADER}[*] {message}{bcolors.ENDC}")


def print_bold(message: str) -> None:
    print(f"{bcolors.BOLD}[*] {message}{bcolors.ENDC}")


def print_underline(message: str) -> None:
    print(f"{bcolors.UNDERLINE}[*] {message}{bcolors.ENDC}")
