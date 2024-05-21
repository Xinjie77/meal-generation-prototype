import React from "react";
import "./Meal.css";
import axios from "axios";

const MealTable = ({ meals, formData, updateMeal }) => {
  const groupByDay = (meals) => {
    if (!meals || meals.length === 0) {
      return {}; // Return an empty object if meals is undefined or empty
    }

    return meals.reduce((acc, meal) => {
      if (!acc[meal.day]) {
        acc[meal.day] = [];
      }
      acc[meal.day].push(meal);
      return acc;
    }, {});
  };

  const handleSwap = async (meal) => {
    const otherMeals = meals
      .filter(
        (ex) => ex.day === meal.day && ex.meal !== meal.meal
      )
      .map((ex) => ex.meal);

    try {
      const response = await axios.post("http://localhost:8080/swap-meal", {
        meal: meal,
        otherMeals: otherMeals,
        dietPreference: formData.dietPreference,
        allergies: formData.allergies,
      });
      console.log("Swap response data:", response.data); // Log the response data
      const oldMealIndigredient = response.data.oldMealIndigredient;
      const newMeal = response.data.newMeal;
      updateMeal(meal.day, oldMealIndigredient, newMeal);
    } catch (error) {
      console.error("Error swapping meal:", error);
    }
  };

  const groupedMeals = groupByDay(meals);

  return (
    <div className="meal-table-container">
      <table className="meal-table">
        <thead>
          <tr>
            <th>Day</th>
            <th>Meal</th>
            <th>Ingredients</th>
            <th>Instructions</th>
            <th>Nutrition(Approx.)</th>
            <th>Action</th>
          </tr>
        </thead>
        <tbody>
          {Object.keys(groupedMeals).map((day, index) => (
            <React.Fragment key={index}>
              <tr className="day-header">
                <td colSpan="6">{day}</td>
              </tr>
              {groupedMeals[day].map((meal, idx) => (
                <tr key={idx}>
                  <td>{meal.mealType}</td>
                  <td>{meal.ingredients}</td>
                  <td>{meal.instructions}</td>
                  <td>{meal.nutrition}</td>
                  <td>
                    <button
                      className="swap-button"
                      onClick={() => handleSwap(meal)}
                    >
                      Swap
                    </button>
                  </td>
                </tr>
              ))}
            </React.Fragment>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default MealTable;
