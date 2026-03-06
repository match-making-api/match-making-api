import unittest

class TestPrizeDistribution(unittest.TestCase):
    def test_process_match_results(self):
        match_results = {
            'match_id': 123,
            'tournament_id': 456,
            'prize_pool': 1000,
            'distribution_rules': 'winner-take-all',
            'winners': [
                {'player_id': 1, 'performance': 'first'},
            ]
        }
        try:
            process_match_results(match_results)
            self.assertTrue(True)
        except Exception as e:
            self.fail(f"Exception occurred: {e}")

    def test_duplicate_payout_event(self):
        def mock_is_payout_event_exists(match_id):
            return True
        global is_payout_event_exists
        is_payout_event_exists = mock_is_payout_event_exists
        match_results = {
            'match_id': 123,
            'tournament_id': 456,
            'prize_pool': 1000,
            'distribution_rules': 'winner-take-all',
            'winners': [
                {'player_id': 1, 'performance': 'first'},
            ]
        }
        with self.assertRaises(ValueError):
            process_match_results(match_results)

if __name__ == '__main__':
    unittest.main()