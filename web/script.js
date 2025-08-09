document.addEventListener('DOMContentLoaded', () => {
    const orderIdInput = document.getElementById('orderId');
    const getOrderBtn = document.getElementById('getOrderBtn');
    const resultDiv = document.getElementById('result');
    const loadingIndicator = document.getElementById('loading');

    function displayOrder(order) {
        if (!order) {
            resultDiv.textContent = 'Order not found';
            return;
        }

        resultDiv.textContent = JSON.stringify(order, null, 2);

    }

    function displayError(error) {
        resultDiv.textContent = `Error: ${error.message || 'Unknown error'}`;
    }

    getOrderBtn.addEventListener('click', async () => {
        const orderId = orderIdInput.value.trim();

        if (!orderId) {
            alert('Please enter Order ID');
            return;
        }

        loadingIndicator.style.display = 'block';
        resultDiv.textContent = '';

        try {
            const response = await fetch(`/order/${orderId}`);

            if (!response.ok) {
                throw new Error(`Server returned ${response.status}`);
            }

            const order = await response.json();
            displayOrder(order);
        } catch (error) {
            displayError(error);
        } finally {
            loadingIndicator.style.display = 'none';
        }
    });

    orderIdInput.addEventListener('keyup', (event) => {
        if (event.key === 'Enter') {
            getOrderBtn.click();
        }
    });
});
