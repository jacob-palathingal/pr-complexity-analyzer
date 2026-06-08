def process_order(order):
    """More branchy order processor — higher complexity."""
    if not order:
        return 0
    if not order.is_valid():
        return 0
    if order.requires_approval():
        if order.approval_pending():
            return None
        if not order.is_approved():
            raise PermissionError("order not approved")
    total = order.total()
    if total > 10000:
        total = apply_large_order_discount(total)
    return total


def apply_large_order_discount(total):
    discount = 0.05
    if total > 50000:
        discount = 0.10
    return total * (1 - discount)


class PaymentHandler:
    def charge(self, card, amount, currency="USD", retry=False):
        """More complex charge — handles retries and currency."""
        if amount <= 0:
            raise ValueError("amount must be positive")
        if currency not in ("USD", "EUR", "GBP"):
            raise ValueError(f"unsupported currency: {currency}")
        try:
            return card.charge(amount, currency)
        except card.DeclinedError:
            if retry:
                return card.charge(amount, currency)
            raise
