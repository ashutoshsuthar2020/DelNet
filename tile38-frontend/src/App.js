import React, { useState, useEffect } from "react";
import { MapContainer, TileLayer, Marker, useMapEvents } from "react-leaflet";
import "leaflet/dist/leaflet.css";
import L from "leaflet";

const storeIcon = new L.Icon({
  iconUrl: "store-icon.png",
  iconSize: [25, 25],
});

const deliveryIcon = new L.Icon({
  iconUrl: "delivery-icon.png",
  iconSize: [25, 25],
});

const driverIcon = new L.Icon({
  iconUrl: "driver-icon.png",
  iconSize: [25, 25],
});

const LocationMarker = ({ setSelectedLocation }) => {
  useMapEvents({
    click(e) {
      setSelectedLocation({ lat: e.latlng.lat, lng: e.latlng.lng });
    },
  });
  return null;
};

const App = () => {
  const [stores, setStores] = useState([]);
  const [deliveries, setDeliveries] = useState([]);
  const [drivers, setDrivers] = useState([]);
  const [selectedLocation, setSelectedLocation] = useState(null);
  const [selectedType, setSelectedType] = useState("store");
  const [id, setId] = useState("");
  const [error, setError] = useState("");
  const [showMapData, setShowMapData] = useState(false);

  useEffect(() => {
    const fetchLocations = async () => {
      try {
        const response = await fetch("http://localhost:8089/locations");
        if (!response.ok) {
          throw new Error(`HTTP error! Status: ${response.status}`);
        }
        const data = await response.json();
        setStores(data.stores || []);
        setDeliveries(data.deliveries || []);
        setDrivers(data.drivers || []);
      } catch (error) {
        console.error("Error fetching locations:", error);
        setError("Failed to load locations. Please check the backend.");
      }
    };
    fetchLocations();
    const interval = setInterval(fetchLocations, 5000);
    return () => clearInterval(interval);
  }, []);

  const handleSubmit = async () => {
    if (!id || !selectedLocation) {
      setError("ID and location are required");
      return;
    }

    if (stores.some((s) => s.id === id) || deliveries.some((d) => d.id === id)) {
      setError("Store and Delivery IDs must be unique");
      return;
    }

    const existingDriver = drivers.find((driver) => driver.id === id);
    if (selectedType === "driver" && existingDriver) {
      try {
        const response = await fetch("http://localhost:8089/updateDriverLocation", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ id, ...selectedLocation }),
        });

        if (!response.ok) {
          throw new Error("Failed to update driver location");
        }

        setDrivers(
          drivers.map((driver) =>
            driver.id === id ? { ...driver, ...selectedLocation } : driver
          )
        );
        setError("");
        return;
      } catch (error) {
        console.error("Error updating driver location:", error);
        setError("Failed to update driver location.");
        return;
      }
    }

    const newEntry = { id, type: selectedType, ...selectedLocation };

    try {
      const response = await fetch("http://localhost:8089/addLocation", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(newEntry),
      });

      if (!response.ok) {
        throw new Error("Failed to save location");
      }

      if (selectedType === "store") setStores([...stores, newEntry]);
      else if (selectedType === "delivery") setDeliveries([...deliveries, newEntry]);
      else if (selectedType === "driver") setDrivers([...drivers, newEntry]);

      setError("");
    } catch (error) {
      console.error("Error saving location:", error);
      setError("Failed to save location.");
    }
  };

  return (
    <div>
      <h1>Map Location Selector</h1>
      <select value={selectedType} onChange={(e) => setSelectedType(e.target.value)}>
        <option value="store">Store</option>
        <option value="delivery">Delivery</option>
        <option value="driver">Driver</option>
      </select>
      <input
        type="text"
        placeholder="Enter ID"
        value={id}
        onChange={(e) => setId(e.target.value)}
      />
      <button onClick={handleSubmit}>Submit Location</button>
      <button onClick={() => setShowMapData(!showMapData)}>Show Map Data</button>
      {error && <p style={{ color: "red" }}>{error}</p>}

      <MapContainer center={[37.7749, -122.4194]} zoom={13} style={{ height: "400px", width: "100%" }}>
        <TileLayer url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png" />
        <LocationMarker setSelectedLocation={setSelectedLocation} />
        {stores.map((store) => (
          <Marker key={store.id} position={[store.lat, store.lng]} icon={storeIcon} />
        ))}
        {deliveries.map((delivery) => (
          <Marker key={delivery.id} position={[delivery.lat, delivery.lng]} icon={deliveryIcon} />
        ))}
        {drivers.map((driver) => (
          <Marker key={driver.id} position={[driver.lat, driver.lng]} icon={driverIcon} />
        ))}
      </MapContainer>

      {showMapData && (
        <div>
          <h2>Saved Locations</h2>
          <ul>
            {stores.map((store) => (
              <li key={store.id}>
                🛒 Store {store.id}: ({store.lat}, {store.lng}) 
                <button onClick={async () => {
                  try {
                    const response = await fetch(`http://localhost:8089/stores`, {
                      method: "DELETE",
                      headers: { "Content-Type": "application/json" },
                      body: JSON.stringify({ id: store.id }),
                    });
                    if (!response.ok) throw new Error("Failed to delete store");
                    setStores(stores.filter((s) => s.id !== store.id));
                  } catch (error) {
                    console.error("Error deleting store:", error);
                    setError("Failed to delete store.");
                  }
                }}>Delete</button>
              </li>
            ))}
            {deliveries.map((delivery) => (
              <li key={delivery.id}>
                📦 Delivery {delivery.id}: ({delivery.lat}, {delivery.lng}) 
                <button onClick={async () => {
                  try {
                    const response = await fetch(`http://localhost:8089/deliveries`, {
                      method: "DELETE",
                      headers: { "Content-Type": "application/json" },
                      body: JSON.stringify({ id: delivery.id }),
                    });
                    if (!response.ok) throw new Error("Failed to delete delivery");
                    setDeliveries(deliveries.filter((d) => d.id !== delivery.id));
                  } catch (error) {
                    console.error("Error deleting delivery:", error);
                    setError("Failed to delete delivery.");
                  }
                }}>Delete</button>
              </li>
            ))}
            {drivers.map((driver) => (
              <li key={driver.id}>
                🚗 Driver {driver.id}: ({driver.lat}, {driver.lng}) 
                <button onClick={async () => {
                  try {
                    const response = await fetch(`http://localhost:8089/drivers`, {
                      method: "DELETE",
                      headers: { "Content-Type": "application/json" },
                      body: JSON.stringify({ id: driver.id }),
                    });
                    if (!response.ok) throw new Error("Failed to delete driver");
                    setDrivers(drivers.filter((d) => d.id !== driver.id));
                  } catch (error) {
                    console.error("Error deleting driver:", error);
                    setError("Failed to delete driver.");
                  }
                }}>Delete</button>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
};

export default App;