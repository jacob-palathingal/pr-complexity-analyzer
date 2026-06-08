# Fixture: simple Python file before the PR change.
# Used by radon_test.go to validate complexity scoring.


def process_order(order):
    """Simple order processor — low complexity."""
    if order.is_valid():
        return order.total()
    return 0


class PaymentHandler:
    def charge(self, card, amount):
        """Straightforward charge — complexity 2."""
        if amount <= 0:
            raise ValueError("amount must be positive")
        return card.charge(amount)
