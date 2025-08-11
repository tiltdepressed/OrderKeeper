document.addEventListener('DOMContentLoaded', () => {
    console.log('DOM fully loaded and parsed');

    const orderIdInput = document.getElementById('orderId');
    const getOrderBtn = document.getElementById('getOrderBtn');
    const resultDiv = document.getElementById('result');
    const loadingIndicator = document.getElementById('loading');

    function displayOrder(order) {
        console.log('Displaying order:', order);
        if (!order || Object.keys(order).length === 0) {
            resultDiv.textContent = 'Order not found';
            resultDiv.classList.remove('error');
            return;
        }
        resultDiv.textContent = JSON.stringify(order, null, 2);
        resultDiv.classList.remove('error');
    }

    function displayError(error) {
        console.error('Error occurred:', error);
        resultDiv.textContent = `Error: ${error.message || 'Unknown error'}`;
        resultDiv.classList.add('error');
    }

    async function fetchOrder(orderId) {
        console.log(`Fetching order for ID: ${orderId}`);
        loadingIndicator.style.display = 'block';
        resultDiv.textContent = '';
        resultDiv.classList.remove('error');

        try {
            const response = await fetch(`/order/${encodeURIComponent(orderId)}`);
            console.log('Response received:', response);

            if (!response.ok) {
                let errorText = await response.text();
                throw new Error(`Server returned ${response.status}${errorText ? ': ' + errorText : ''}`);
            }

            const order = await response.json();
            console.log('Order received:', order);
            displayOrder(order);
        } catch (error) {
            displayError(error);
        } finally {
            loadingIndicator.style.display = 'none';
            console.log('Fetch operation completed');
        }
    }

    getOrderBtn.addEventListener('click', () => {
        console.log('Get Order button clicked');
        const orderId = orderIdInput.value.trim();
        console.log('Order ID entered:', orderId);

        if (!orderId) {
            alert('Please enter Order ID');
            console.log('No Order ID entered, alert shown');
            return;
        }

        fetchOrder(orderId);
    });

    orderIdInput.addEventListener('keyup', (event) => {
        if (event.key === 'Enter') {
            console.log('Enter key pressed in orderId input');
            getOrderBtn.click();
        }
    });
});

