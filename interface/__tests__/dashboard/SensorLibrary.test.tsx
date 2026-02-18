import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor, act, fireEvent } from '@testing-library/react';
import SensorLibrary from '@/components/dashboard/SensorLibrary';
import { useCortexStore } from '@/store/useCortexStore';
import { mockFetch } from '../setup';

const SENSOR_RESPONSE = {
    sensors: [
        { id: 'sensor-gmail-inbox', type: 'email', status: 'online', last_seen: new Date().toISOString(), label: 'Gmail Inbox' },
        { id: 'sensor-gmail-sent', type: 'email', status: 'online', last_seen: new Date().toISOString(), label: 'Gmail Sent' },
        { id: 'sensor-weather-local', type: 'weather', status: 'online', last_seen: new Date().toISOString(), label: 'Local Weather' },
        { id: 'sensor-weather-forecast', type: 'weather', status: 'degraded', last_seen: new Date().toISOString(), label: 'Forecast' },
        { id: 'sensor-pg-primary', type: 'database', status: 'online', last_seen: new Date().toISOString(), label: 'PostgreSQL' },
    ],
};

describe('SensorLibrary', () => {
    beforeEach(() => {
        vi.useFakeTimers({ shouldAdvanceTime: true });
        useCortexStore.setState({
            sensorFeeds: [],
            isFetchingSensors: false,
            subscribedSensorGroups: [],
        });
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    it('renders sensor library container', () => {
        mockFetch.mockResolvedValue({ ok: true, json: async () => ({ sensors: [] }) });
        render(<SensorLibrary />);
        expect(screen.getByTestId('sensor-library')).toBeDefined();
    });

    it('shows "No sensor groups" when API returns empty', async () => {
        mockFetch.mockResolvedValue({ ok: true, json: async () => ({ sensors: [] }) });

        render(<SensorLibrary />);

        await waitFor(() => {
            expect(screen.getByText('No sensor groups')).toBeDefined();
        });
    });

    it('renders sensor groups from API data', async () => {
        mockFetch.mockResolvedValue({ ok: true, json: async () => SENSOR_RESPONSE });

        render(<SensorLibrary />);

        await waitFor(() => {
            expect(screen.getByText('database')).toBeDefined();
            expect(screen.getByText('email')).toBeDefined();
            expect(screen.getByText('weather')).toBeDefined();
        });
    });

    it('shows online/total count per group', async () => {
        mockFetch.mockResolvedValue({ ok: true, json: async () => SENSOR_RESPONSE });

        render(<SensorLibrary />);

        await waitFor(() => {
            // email: 2/2 online, weather: 1/2 online, database: 1/1 online
            expect(screen.getByText('2/2 online')).toBeDefined();
            expect(screen.getByText('1/2 online')).toBeDefined();
            expect(screen.getByText('1/1 online')).toBeDefined();
        });
    });

    it('shows total online count in header', async () => {
        mockFetch.mockResolvedValue({ ok: true, json: async () => SENSOR_RESPONSE });

        render(<SensorLibrary />);

        await waitFor(() => {
            expect(screen.getByText('4/5')).toBeDefined(); // 4 online out of 5
        });
    });

    it('fetches sensors on mount', async () => {
        mockFetch.mockResolvedValue({ ok: true, json: async () => SENSOR_RESPONSE });

        render(<SensorLibrary />);

        await waitFor(() => {
            expect(mockFetch).toHaveBeenCalledWith('/api/v1/sensors');
        });
    });

    it('expands group on click to show individual sensors', async () => {
        mockFetch.mockResolvedValue({ ok: true, json: async () => SENSOR_RESPONSE });

        render(<SensorLibrary />);

        await waitFor(() => {
            expect(screen.getByText('email')).toBeDefined();
        });

        // Click the email group to subscribe/expand
        act(() => {
            fireEvent.click(screen.getByText('email'));
        });

        await waitFor(() => {
            expect(screen.getByText('Gmail Inbox')).toBeDefined();
            expect(screen.getByText('Gmail Sent')).toBeDefined();
        });
    });

    it('shows SUB count in header when groups are subscribed', async () => {
        mockFetch.mockResolvedValue({ ok: true, json: async () => SENSOR_RESPONSE });

        render(<SensorLibrary />);

        await waitFor(() => {
            expect(screen.getByText('email')).toBeDefined();
        });

        act(() => {
            fireEvent.click(screen.getByText('email'));
        });

        await waitFor(() => {
            expect(screen.getByText('1 SUB')).toBeDefined();
        });
    });

    it('shows loading skeleton while fetching', () => {
        useCortexStore.setState({ isFetchingSensors: true, sensorFeeds: [] });
        mockFetch.mockReturnValue(new Promise(() => {}));

        const { container } = render(<SensorLibrary />);

        expect(container.querySelectorAll('.animate-pulse').length).toBeGreaterThanOrEqual(1);
    });

    it('handles API failure gracefully', async () => {
        mockFetch.mockRejectedValue(new Error('Network error'));

        render(<SensorLibrary />);

        await waitFor(() => {
            expect(screen.getByText('No sensor groups')).toBeDefined();
        });
    });
});
