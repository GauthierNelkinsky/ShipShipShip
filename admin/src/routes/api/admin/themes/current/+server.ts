import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

const BACKEND_URL = 'http://localhost:8080';

export const GET: RequestHandler = async ({ cookies }) => {
	try {
		// Get auth token from cookies
		const token = cookies.get('auth_token');
		if (!token) {
			return json({ error: 'Authentication required' }, { status: 401 });
		}

		// Forward request to Go backend
		const response = await fetch(`${BACKEND_URL}/api/admin/themes/current`, {
			method: 'GET',
			headers: {
				'Authorization': `Bearer ${token}`
			}
		});

		const data = await response.json();

		if (!response.ok) {
			return json(data, { status: response.status });
		}

		return json(data);
	} catch (error) {
		console.error('Get current theme error:', error);
		return json({ error: 'Failed to get current theme' }, { status: 500 });
	}
};
