import io
import json
import unittest
from unittest.mock import patch

import openai_response


class OpenAIResponseTest(unittest.TestCase):
    def test_build_payload_uses_requested_model_and_effort(self) -> None:
        payload = openai_response.build_payload("Write this", "gpt-5.6-sol", "medium", 8000)

        self.assertEqual(payload["model"], "gpt-5.6-sol")
        self.assertEqual(payload["reasoning"], {"effort": "medium"})
        self.assertEqual(payload["input"], "Write this")
        self.assertEqual(payload["max_output_tokens"], 8000)

    def test_extract_output_text_from_responses_envelope(self) -> None:
        response = {
            "output": [
                {
                    "type": "message",
                    "content": [
                        {"type": "output_text", "text": '{"post_text":"Clear copy"}'}
                    ],
                }
            ]
        }

        self.assertEqual(
            openai_response.extract_output_text(response),
            '{"post_text":"Clear copy"}',
        )

    def test_extract_output_text_rejects_empty_response(self) -> None:
        with self.assertRaisesRegex(ValueError, "no output text"):
            openai_response.extract_output_text({"output": []})

    def test_response_url_accepts_api_root_or_v1_root(self) -> None:
        expected = "https://api.openai.com/v1/responses"
        self.assertEqual(openai_response.response_url("https://api.openai.com"), expected)
        self.assertEqual(openai_response.response_url("https://api.openai.com/v1/"), expected)

    @patch("openai_response.urllib.request.urlopen")
    def test_create_response_posts_responses_payload(self, urlopen) -> None:
        envelope = {
            "output": [
                {
                    "type": "message",
                    "content": [{"type": "output_text", "text": "finished"}],
                }
            ]
        }
        urlopen.return_value = io.BytesIO(json.dumps(envelope).encode("utf-8"))

        result = openai_response.create_response(
            prompt="Draft a post",
            api_key="test-key",
            model="gpt-5.6-sol",
            reasoning_effort="medium",
            max_output_tokens=16000,
            timeout_seconds=30,
            base_url="https://api.openai.com",
        )

        self.assertEqual(result, "finished")
        request = urlopen.call_args.args[0]
        payload = json.loads(request.data)
        self.assertEqual(payload["model"], "gpt-5.6-sol")
        self.assertEqual(payload["reasoning"], {"effort": "medium"})
        self.assertEqual(request.full_url, "https://api.openai.com/v1/responses")


if __name__ == "__main__":
    unittest.main()
