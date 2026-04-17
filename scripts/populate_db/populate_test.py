import base64
import random

import utils.utils as utils
from models.assessment import Assessment
from models.base import Base
from models.category import Category
from models.customer import Customer
from models.poc import Poc, PocData
from models.target import Target
from models.user import User
from models.vulnerability import Vulnerability
from requests import Session


def populate_test(session: Session, base_url: str, username: str, password: str) -> str:
    # initialize Base with session and base_url
    Base.base_url = base_url
    Base.session = session

    image = utils.rand_image()

    # login as admin
    user = User(username=username, password=password)
    if not user.login():
        raise Exception(f"Failed to login as {user.username}")

    # create users
    users = create_users(10)

    # create categories
    categories = create_categories(5)

    # create customers
    customers = create_customers(5)

    # create targets for each customer
    for customer in customers:
        targets = create_targets(customer.id, 5)

    # create assessments
    assessments = []
    for customer in customers:
        assessments += create_assessments(customer, 3)

    for assessment in assessments:
        vulnerabilities = create_vulnerabilities(assessment, categories, 5)

        for vulnerability in vulnerabilities:
            pocs = create_pocs(vulnerability, 7)


def create_users(n: int) -> list:
    """
    Create n users with random usernames and a fixed password.

    :param n: Number of users to create
    :return: List of User objects
    """

    users = []
    for _ in range(n):
        user = User()

        id, error = user.add()
        if error != "":
            raise Exception(f"Failed to register user {user.username}: {error}")
        users.append(user)
        utils.print_success(
            f"Registered user {user.username} {user.password} with id {id}"
        )

    return users


def create_categories(n: int) -> list:
    """
    Create n categories with random names and generic remediation.

    :param n: Number of categories to create
    :return: List of Category objects
    """

    categories = []
    for identifier in utils.rand_category_identifiers(n):
        category = Category(identifier=identifier)
        id, error = category.add()
        if error != "":
            raise Exception(f"Failed to register category {category.name}: {error}")
        categories.append(category)
        utils.print_success(f"Registered category {category.name} with id {id}")

    return categories


def create_customers(n: int) -> list:
    """
    Create n customers with random names.

    :param n: Number of customers to create
    :return: List of Customer objects
    """

    customers = []
    for _ in range(n):
        customer = Customer()
        id, error = customer.add()
        if error != "":
            raise Exception(f"Failed to register customer {customer.name}: {error}")
        customers.append(customer)
        utils.print_success(f"Registered customer {customer.name} with id {id}")

    return customers


def create_targets(customer_id: str, n: int) -> list:
    """
    Create n targets for a given customer.

    :param customer: Customer object to associate targets with
    :param n: Number of targets to create
    :return: List of Target objects
    """

    targets = []
    for _ in range(n):
        target = Target(
            customer_id=customer_id,
        )
        id, error = target.add()
        if error != "":
            raise Exception(f"Failed to register target {target.fqdn}: {error}")
        targets.append(target)
        utils.print_success(f"Registered target {target.fqdn} with id {id}")

    return targets


def create_assessments(customer: Customer, n: int) -> list:
    """
    Create n assessments.

    :param customer_id: Customer ID to associate assessments with
    :param n: Number of assessments to create
    :return: List of Assessment objects
    """

    assessments = []
    for _ in range(n):
        targets = customer.getTargets()
        assessment = Assessment(
            customer_id=customer.id,
            targets=[
                x["id"]
                for x in random.sample(targets, k=random.randint(1, len(targets)))
            ],
        )
        id, error = assessment.add()
        if error != "":
            raise Exception(f"Failed to register assessment {assessment.name}: {error}")
        assessments.append(assessment)
        utils.print_success(f"Registered assessment {assessment.name} with id {id}")

    return assessments


def create_vulnerabilities(assessment: Assessment, categories: list, n: int) -> list:
    """
    Create n vulnerabilities for each assessment.

    :param assessments: List of Assessment objects
    :param categories: List of Category objects
    :param n: Number of vulnerabilities to create for each assessment
    :return: List of Vulnerability objects
    """

    vulnerabilities = []
    for _ in range(n):
        category = random.choice(categories)
        vulnerability = Vulnerability(
            category=category,
            target_id=random.choice(assessment.targets),
            assessment=assessment,
        )
        id, error = vulnerability.add()
        if error != "":
            raise Exception(
                f"Failed to register vulnerability {vulnerability.detailed_title}: {error}"
            )
        vulnerabilities.append(vulnerability)
        utils.print_success(
            f"Registered vulnerability {vulnerability.detailed_title} with id {id}"
        )

    return vulnerabilities


def create_pocs(vulnerability: Vulnerability, n: int) -> list:
    """
    Create n POCs for a given vulnerability.

    :param vulnerability: Vulnerability object to associate POCs with
    :param n: Number of POCs to create
    :return: List of POC objects
    """

    pocs = []
    for _ in range(n):
        pocs_data = []
        for i in range(7):
            poc_data = PocData(
                index=i,
            )
            pocs_data.append(poc_data)
        poc = Poc(poc_data=pocs_data, vulnerability_id=vulnerability.id)
        id, error = poc.add()
        if error != "":
            raise Exception(f"Failed to register POC: {error}")
        pocs.append(poc)
        utils.print_success(
            f"Registered POC for Vulnerability with id {vulnerability.id}"
        )

    return pocs
