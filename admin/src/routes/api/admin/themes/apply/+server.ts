import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

const BACKEND_URL = 'http://localhost:8080';

export const POST: RequestHandler = async ({ request, cookies }) => {
	try {
		const body = await request.json();

		// Get auth token from cookies
		const token = cookies.get('auth_token');
		if (!token) {
			return json({ error: 'Authentication required' }, { status: 401 });
		}

		// Forward request to Go backend
		const response = await fetch(`${BACKEND_URL}/api/admin/themes/apply`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
				'Authorization': `Bearer ${token}`
			},
			body: JSON.stringify(body)
		});

		const data = await response.json();

		if (!response.ok) {
			return json(data, { status: response.status });
		}

		return json(data);
	} catch (error) {
		console.error('Theme apply error:', error);
		return json({ error: 'Failed to apply theme' }, { status: 500 });
	}
};
