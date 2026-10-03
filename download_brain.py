# /// script
# requires-python = ">=3.10"
# dependencies = [
#     "neuprint-python",
#     "python-dotenv",
#     "pandas",
# ]
# ///

import logging
import os
import sys
from pathlib import Path
from tempfile import TemporaryDirectory

from dotenv import load_dotenv
from neuprint import Client

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")
load_dotenv()

# Configuration
SERVER = "neuprint.janelia.org"
DATASET = "male-cns:v1.0"
MIN_SYNAPSE_WEIGHT = 5

BASE_DIR = Path(__file__).resolve().parent
OUTPUT_DIR = BASE_DIR / "data"
OUTPUT_FILE = OUTPUT_DIR / "manc_synapses.csv"


def fetch_connectome() -> Path:
    token = os.getenv("NEUPRINT_TOKEN")
    if not token:
        logging.error("Missing 'NEUPRINT_TOKEN'. Please set it in your .env file.")
        sys.exit(1)

    logging.info(f"Connecting to NeuPrint ({SERVER} | {DATASET})...")
    client = Client(SERVER, dataset=DATASET, token=token)

    query = f"""
        MATCH (n:Neuron)-[e:ConnectsTo]->(m:Neuron)
        WHERE e.weight >= {MIN_SYNAPSE_WEIGHT}
        RETURN n.bodyId AS pre_bodyId, m.bodyId AS post_bodyId, e.weight AS weight
    """

    logging.info(f"Fetching connectome data (weight >= {MIN_SYNAPSE_WEIGHT})...")
    results = client.fetch_custom(query)
    logging.info(f"Retrieved {len(results):,} synaptic connections.")

    # Replace the previous dataset only after a complete CSV write.
    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)
    with TemporaryDirectory(dir=OUTPUT_DIR, prefix=".download-") as temp_dir:
        temp_file = Path(temp_dir) / OUTPUT_FILE.name
        results.to_csv(temp_file, index=False)
        temp_file.replace(OUTPUT_FILE)
    logging.info(f"Dataset successfully saved to: {OUTPUT_FILE}")

    return OUTPUT_FILE


if __name__ == "__main__":
    fetch_connectome()
